package iptables

import (
	"bufio"
	"bytes"
	"errors"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"sync"

	"github.com/deckarep/golang-set"
	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/net"
	"github.com/luizbafilho/fusis/state"
)

var (
	ErrIptablesNotFound = errors.New("Iptables not found")
	ErrIptablesSnat     = errors.New("Iptables: error when inserting SNAT rule")
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
	path, err := exec.LookPath("iptables")
	if err != nil {
		return nil, ErrIptablesNotFound
	}

	return &IptablesMngr{
		config: config,
		path:   path,
	}, nil
}

func (i IptablesMngr) Sync(s state.State) error {
	i.Lock()
	defer i.Unlock()

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

	for r := range rulesToAdd.Iter() {
		rule := r.(SnatRule)
		err := i.addRule(rule)
		if err != nil {
			return err
		}
	}

	for r := range rulesToRemove.Iter() {
		rule := r.(SnatRule)
		err := i.removeRule(rule)
		if err != nil {
			return err
		}
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
		vaddr:    svc.Host,
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
	cmd := exec.Command(i.path, "--wait", "-t", "nat", action, "POSTROUTING", "-m", "ipvs", "--vaddr", r.vaddr+"/32", "--vport", r.vport, "-j", "SNAT", "--to-source", r.toSource)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (i IptablesMngr) getSnatRules() ([]SnatRule, error) {
	out, err := exec.Command(i.path, "--wait", "--list", "-t", "nat").Output()
	if err != nil {
		log.Fatal(err)
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
