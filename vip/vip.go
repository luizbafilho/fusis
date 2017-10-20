package vip

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/deckarep/golang-set"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/net"
	"github.com/luizbafilho/fusis/state"
)

type VipMngr struct {
	iface string
}

type Syncer interface {
	Sync(s state.State) error
}

func New(config *config.BalancerConfig) (Syncer, error) {
	return &VipMngr{
		iface: config.Interfaces.Inbound,
	}, nil
}

func (n VipMngr) Sync(state state.State) error {
	start := time.Now()
	defer func() {
		log.Debugf("[vip] Sync took %v", time.Since(start))
	}()

	log.Debug("[vip] Syncing")
	currentVips, err := n.getCurrentVips()
	if err != nil {
		return err
	}

	stateVips := n.getStateVips(state)

	vipsToAdd := stateVips.Difference(currentVips)
	vipsToRemove := currentVips.Difference(stateVips)

	for v := range vipsToAdd.Iter() {
		vip := v.(string)
		err := net.AddIp(vip+"/32", n.iface)
		if err != nil {
			return fmt.Errorf("[vip] Error adding ip %s: %s", vip, err)
		}
		log.Debugf("[vip] Added: %s/32 to interface: %s", vip, n.iface)
	}

	for v := range vipsToRemove.Iter() {
		vip := v.(string)
		err := net.DelIp(vip+"/32", n.iface)
		if err != nil {
			return fmt.Errorf("[vip] Error deleting ip %s: %s", vip, err)
		}
		log.Debugf("[vip] Removed: %s/32 from interface: %s", vip, n.iface)
	}

	return nil
}

func (n VipMngr) getCurrentVips() (mapset.Set, error) {
	vips, err := net.GetFusisVipsIps(n.iface)
	if err != nil {
		return nil, err
	}

	set := mapset.NewSet()
	for _, v := range vips {
		set.Add(v)
	}

	return set, nil
}

func (n VipMngr) getStateVips(state state.State) mapset.Set {
	vips := []string{}

	svcs := state.GetServices()
	for _, s := range svcs {
		vips = append(vips, s.Address)
	}

	set := mapset.NewSet()
	for _, v := range vips {
		set.Add(v)
	}

	return set
}
