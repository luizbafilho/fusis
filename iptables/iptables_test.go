package iptables

import (
	"os"
	"testing"
	"os/exec"

	"github.com/luizbafilho/fusis/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/net"
	"github.com/luizbafilho/fusis/state/mocks"
	"github.com/stretchr/testify/assert"
)

/** TestIptablesSync checks if iptable rules are beeing synced with stored state */
func TestIptablesSync(t *testing.T) {
	if os.Getenv("TRAVIS") == "true" {
		t.Skip("Skipping test because travis-ci do not allow iptables")
	}

	s1 := types.Service{
		Name:     "test",
		Address:  "10.0.1.1",
		Port:     80,
		Mode:     "nat",
		Protocol: "tcp",
	}

	s2 := types.Service{
		Name:     "test2",
		Address:  "10.0.1.2",
		Port:     80,
		Protocol: "tcp",
		Mode:     "nat",
	}

	state := &mocks.State{}
	state.On("GetServices").Return([]types.Service{s1, s2})

	toSource, err := net.GetIpByInterface("lo")
	assert.Nil(t, err)

	rule2 := SnatRule{
		vaddr:    "10.0.1.2",
		vport:    "80",
		toSource: toSource,
	}
	rule3 := SnatRule{
		vaddr:    "10.0.1.3",
		vport:    "80",
		toSource: toSource,
	}

	iptablesMngr, err := New(defaultConfig())
	assert.Nil(t, err)

	err = iptablesMngr.addRule(rule2)
	assert.Nil(t, err)
	err = iptablesMngr.addRule(rule3)
	assert.Nil(t, err)

	err = iptablesMngr.Sync(state)
	assert.Nil(t, err)

	rules, err := iptablesMngr.getKernelRulesSet()
	assert.Nil(t, err)

	rule1, err := iptablesMngr.serviceToSnatRule(s1)
	assert.Nil(t, err)

	assert.Equal(t, rules.Contains(rule2, *rule1), true)

	cleanupRules(t, iptablesMngr)
}

func TestGetSnatRules(t *testing.T) {
	iptablesMngr, err := New(defaultConfig())
	assert.Nil(t, err)

	// ensure there is a FUSIS chain, we don't care about the return code
	_ = exec.Command(iptablesMngr.path, "--wait", "-t", "nat", "-N", "FUSIS").Run()
	// ensure the FUSIS chain is empty, flushed
	err = exec.Command(iptablesMngr.path, "--wait", "-t", "nat", "-F", "FUSIS").Run()
	assert.Nil(t, err)

	// assert getSnatRules return an empty list
	kernelRules, err := iptablesMngr.getSnatRules()
	assert.Nil(t, err)
	assert.Equal(t, kernelRules, []SnatRule{})

	// add snat rule
	err = exec.Command(iptablesMngr.path, "--wait", "-t", "nat", "-A", "FUSIS", "-m", "ipvs", "--vaddr", "192.168.1.4/32", "--vport", "7004", "-j", "SNAT", "--to-source 10.0.3.4").Run()
	assert.Nil(t, err)

	// assert getSnatRules return first rule
	kernelRules, err = iptablesMngr.getSnatRules()
	assert.Nil(t, err)
	assert.Nil(t, kernelRules)
}

/** get iptables rules and removes them */
func cleanupRules(t *testing.T, iptablesMngr *IptablesMngr) {
	rules, err := iptablesMngr.getSnatRules()
	assert.Nil(t, err)

	for _, r := range rules {
		err := iptablesMngr.removeRule(r)
		assert.Nil(t, err)
	}
}

/** mocks the config.BalencerConfig */
func defaultConfig() *config.BalancerConfig {
	return &config.BalancerConfig{
		Interfaces: config.Interfaces{
			Inbound:  "lo",
			Outbound: "lo",
		},
		Name: "Test",
		Ipam: config.Ipam{
			Ranges: []string{"192.168.0.0/28"},
		},
	}
}
