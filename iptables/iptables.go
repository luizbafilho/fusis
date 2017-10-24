package iptables

import (
	"bufio"
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/deckarep/golang-set"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/net"
	"github.com/luizbafilho/fusis/state"
	"github.com/luizbafilho/fusis/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	ErrIptablesNotFound = errors.New("[iptables] Binary not found")
	ErrIptablesSnat     = errors.New("[iptables] Error when inserting SNAT rule")
	ErrIptablesRule     = errors.New("[iptables] Error adding rule")
)

// Defines iptables actions
var (
	ADD = "-A"
	DEL = "-D"
)

type Syncer interface {
	Sync(state state.State) error
}

type IptablesMngr struct {
	sync.Mutex
	config *config.BalancerConfig
	path   string
}

type SnatRule struct {
	vaddr    string
	vport    string
	toSource string
}

func New(config *config.BalancerConfig) (*IptablesMngr, error) {
	// look for iptables
	path, err := exec.LookPath("iptables")
	if err != nil {
		return nil, ErrIptablesNotFound
	}

	// create FUSIS iptables chain, ignore error in case FUSIS chain already exists
	_ = exec.Command(path, "-t", "nat", "-N", "FUSIS").Run()

	// if FUSIS chain is not present in POSTROUTING
	err = exec.Command(path, "-t", "nat", "-C", "POSTROUTING", "-j", "FUSIS").Run()
	if err != nil {
		// add FUSIS chain
		err = exec.Command(path, "-t", "nat", "-A", "POSTROUTING", "-j", "FUSIS").Run()
		if err != nil {
			return nil, ErrIptablesRule
		}
	}

	if err := enableContrack(); err != nil {
		return nil, err
	}

	return &IptablesMngr{
		config: config,
		path:   path,
	}, nil
}

// Sync syncs all iptables rules
func (i IptablesMngr) Sync(s state.State) error {
	start := time.Now()
	defer func() {
		log.Debugf("[iptables ] Sync took %v", time.Since(start))
	}()

	log.Debug("[iptables] Syncing")

	stateSet, err := i.getStateRulesSet(s)
	if err != nil {
		return err
	}

	kernelSet, err := i.getKernelRulesSet()
	if err != nil {
		return err
	}

	rulesToAdd := stateSet.Difference(kernelSet)
	rulesToRemove := kernelSet.Difference(stateSet)

	// Adding missing rules
	for r := range rulesToAdd.Iter() {
		rule := r.(SnatRule)
		err := i.addRule(rule)
		if err != nil {
			return err
		}
		log.Debugf("[iptables] Added rule: %#v", rule)
	}

	// Cleaning rules
	for r := range rulesToRemove.Iter() {
		rule := r.(SnatRule)
		err := i.removeRule(rule)
		if err != nil {
			return err
		}
		log.Debugf("[iptables] Removed rule: %#v", rule)
	}

	return nil
}

func (i IptablesMngr) getKernelRulesSet() (mapset.Set, error) {
	kernelRules, err := i.getSnatRules()
	if err != nil {
		return nil, err
	}

	kernelSet := mapset.NewSet()
	for _, r := range kernelRules {
		kernelSet.Add(r)
	}

	return kernelSet, nil

}

func (i IptablesMngr) getStateRulesSet(s state.State) (mapset.Set, error) {
	stateRules, err := i.convertServicesToSnatRules(i.getNatServices(s))
	if err != nil {
		return nil, err
	}

	stateSet := mapset.NewSet()
	for _, r := range stateRules {
		stateSet.Add(r)
	}

	return stateSet, nil
}

func (i IptablesMngr) convertServicesToSnatRules(svcs []types.Service) ([]SnatRule, error) {
	rules := []SnatRule{}
	for _, s := range svcs {

		r, err := i.serviceToSnatRule(s)
		if err != nil {
			return rules, err
		}

		rules = append(rules, *r)
	}

	return rules, nil
}

func (i IptablesMngr) getNatServices(s state.State) []types.Service {
	natServices := []types.Service{}

	for _, svc := range s.GetServices() {
		if svc.IsNat() {
			natServices = append(natServices, svc)
		}
	}

	return natServices
}

func (i IptablesMngr) serviceToSnatRule(svc types.Service) (*SnatRule, error) {
	privateIp, err := net.GetIpByInterface(i.config.Interfaces.Outbound)
	if err != nil {
		return nil, err
	}

	rule := &SnatRule{
		vaddr:    svc.Address,
		vport:    strconv.Itoa(int(svc.Port)),
		toSource: privateIp,
	}

	return rule, nil
}

func (i IptablesMngr) addRule(r SnatRule) error {
	return i.execIptablesCommand(ADD, r)
}

func (i IptablesMngr) removeRule(r SnatRule) error {
	return i.execIptablesCommand(DEL, r)
}

func (i IptablesMngr) execIptablesCommand(action string, r SnatRule) error {
	i.Lock()
	defer i.Unlock()
	cmd := exec.Command(i.path, "-t", "nat", action, "FUSIS", "-m", "ipvs", "--vaddr", r.vaddr+"/32", "--vport", r.vport, "-j", "SNAT", "--to-source", r.toSource)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (i IptablesMngr) getSnatRules() ([]SnatRule, error) {
	i.Lock()
	defer i.Unlock()
	out, err := exec.Command(i.path, "--list", "FUSIS", "-t", "nat").Output()
	if err != nil {
		log.Fatal("[iptables] Error executing iptables ", err)
	}

	r, _ := regexp.Compile(`vaddr\s([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})\svport\s(\d+)\sto:([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})`)
	scanner := bufio.NewScanner(bytes.NewReader(out))

	rules := []SnatRule{}
	for scanner.Scan() {
		matches := r.FindStringSubmatch(scanner.Text())
		if len(matches) == 0 {
			continue
		}

		rules = append(rules, SnatRule{
			vaddr:    matches[1],
			vport:    matches[2],
			toSource: matches[3],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func enableContrack() error {
	if err := net.SetSysctl("net.ipv4.vs.conntrack", "1"); err != nil {
		return errors.Wrap(err, "enabling Contrack failed")
	}

	return nil
}
