package iptables

import (
	"os"
	"os/exec"
	"testing"

	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/net"
	"github.com/luizbafilho/fusis/state/mocks"
	"github.com/luizbafilho/fusis/types"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	if os.Getenv("TRAVIS") == "true" {
		t.Skip("Skipping test because travis-ci do not allow iptables")
	}

	// discover iptables binary path
	path, err := exec.LookPath("iptables")
	assert.Nil(t, err)

	// remove fusis chain from POSTROUTING, ignore if fusis chain is not there
	_ = exec.Command(path, "--wait", "-t", "nat", "-D", "POSTROUTING", "-j", "FUSIS").Run()

	// check if FUSIS chain exists
	err = exec.Command(path, "--wait", "-t", "nat", "-L", "FUSIS").Run()
	// if FUSIS chain exists
	if err == nil {
		// ensure fusis chain is empty
		err = exec.Command(path, "--wait", "-t", "nat", "-F", "FUSIS").Run()
		assert.Nil(t, err)
		// delete fusis chain
		err = exec.Command(path, "--wait", "-t", "nat", "-X", "FUSIS").Run()
		assert.Nil(t, err)
	}

	// create iptablesMngr from mocked config
	iptablesMngr, err := New(defaultConfig())
	assert.Nil(t, err)

	// check if FUSIS chain exists
	err = exec.Command(iptablesMngr.path, "--wait", "-t", "nat", "-L", "FUSIS").Run()
	assert.Nil(t, err)

	// check if FUSIS chain is present in POSTROUTING
	err = exec.Command(iptablesMngr.path, "--wait", "-t", "nat", "-C", "POSTROUTING", "-j", "FUSIS").Run()
	assert.Nil(t, err)
}

/** TestIptablesSync checks if iptable rules are beeing synced with stored state */
func TestIptablesSync(t *testing.T) {
	if os.Getenv("TRAVIS") == "true" {
		t.Skip("Skipping test because travis-ci do not allow iptables")
	}

	// create iptablesMngr based on mocked config
	iptablesMngr, err := New(defaultConfig())
	assert.Nil(t, err)

	// ensure the FUSIS chain is empty, flushed
	err = exec.Command(iptablesMngr.path, "--wait", "-t", "nat", "-F", "FUSIS").Run()
	assert.Nil(t, err)

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
}

func TestAddRule(t *testing.T) {
	if os.Getenv("TRAVIS") == "true" {
		t.Skip("Skipping test because travis-ci do not allow iptables")
	}

	// crete iptablesMngr from mocked config
	iptablesMngr, err := New(defaultConfig())
	assert.Nil(t, err)

	// ensure the FUSIS chain is empty, flushed
	err = exec.Command(iptablesMngr.path, "--wait", "-t", "nat", "-F", "FUSIS").Run()
	assert.Nil(t, err)

	// get current lo interface
	toSource, err := net.GetIpByInterface("lo")
	assert.Nil(t, err)

	// mock rule
	rule := SnatRule{
		vaddr:    "10.0.1.1",
		vport:    "80",
		toSource: toSource,
	}

	// call iptables to add rule
	iptablesMngr.addRule(rule)

	// check if rule was added
	err = exec.Command(iptablesMngr.path, "--wait", "-t", "nat", "-C", "FUSIS", "-m", "ipvs", "--vaddr", "10.0.1.1/32", "--vport", "80", "-j", "SNAT", "--to-source", toSource).Run()
	assert.Nil(t, err)
}

func TestRemoveRule(t *testing.T) {
	if os.Getenv("TRAVIS") == "true" {
		t.Skip("Skipping test because travis-ci do not allow iptables")
	}

	// crete iptablesMngr from mocked config
	iptablesMngr, err := New(defaultConfig())
	assert.Nil(t, err)

	// ensure the FUSIS chain is empty, flushed
	err = exec.Command(iptablesMngr.path, "--wait", "-t", "nat", "-F", "FUSIS").Run()
	assert.Nil(t, err)

	// get current lo interface
	toSource, err := net.GetIpByInterface("lo")
	assert.Nil(t, err)

	// mock rule
	rule := SnatRule{
		vaddr:    "10.0.1.1",
		vport:    "80",
		toSource: toSource,
	}

	// add rule
	err = exec.Command(iptablesMngr.path, "--wait", "-t", "nat", "-A", "FUSIS", "-m", "ipvs", "--vaddr", "10.0.1.1/32", "--vport", "80", "-j", "SNAT", "--to-source", toSource).Run()
	assert.Nil(t, err)

	// call iptables to remove rule
	iptablesMngr.removeRule(rule)

	// check using iptables
	err = exec.Command(iptablesMngr.path, "--wait", "-t", "nat", "-C", "FUSIS", "-m", "ipvs", "--vaddr", "10.0.1.1/32", "--vport", "80", "-j", "SNAT", "--to-source", toSource).Run()
	assert.NotNil(t, err)
}

func TestServiceToSnatRule(t *testing.T) {
	if os.Getenv("TRAVIS") == "true" {
		t.Skip("Skipping test because travis-ci do not allow iptables")
	}

	// create iptablesMngr from mocked config
	iptablesMngr, err := New(defaultConfig())
	assert.Nil(t, err)

	// mock service
	s1 := types.Service{
		Name:     "test",
		Address:  "10.0.1.1",
		Port:     80,
		Mode:     "nat",
		Protocol: "tcp",
	}

	// get current lo interface
	toSource, err := net.GetIpByInterface("lo")
	assert.Nil(t, err)

	// convert service to rule
	rule, err := iptablesMngr.serviceToSnatRule(s1)
	assert.Nil(t, err)

	// compare to spected rule
	assert.Equal(t, *rule, SnatRule{
		vaddr:    "10.0.1.1",
		vport:    "80",
		toSource: toSource,
	})
}

func TestGetSnatRules(t *testing.T) {
	if os.Getenv("TRAVIS") == "true" {
		t.Skip("Skipping test because travis-ci do not allow iptables")
	}

	// crete iptablesMngr from mocked config
	iptablesMngr, err := New(defaultConfig())
	assert.Nil(t, err)

	// ensure the FUSIS chain is empty, flushed
	err = exec.Command(iptablesMngr.path, "--wait", "-t", "nat", "-F", "FUSIS").Run()
	assert.Nil(t, err)

	// assert getSnatRules return an empty list
	kernelRules, err := iptablesMngr.getSnatRules()
	assert.Nil(t, err)
	assert.Equal(t, []SnatRule{}, kernelRules)

	// add snat rule to FUSIS chain
	err = exec.Command(iptablesMngr.path, "--wait", "-t", "nat", "-A", "FUSIS", "-m", "ipvs", "--vaddr", "192.168.1.4/32", "--vport", "7004", "-j", "SNAT", "--to-source", "10.0.3.4").Run()
	assert.Nil(t, err)

	// assert getSnatRules return first rule
	kernelRules, err = iptablesMngr.getSnatRules()
	assert.Nil(t, err)
	assert.Equal(t,
		[]SnatRule{
			{
				vaddr:    "192.168.1.4",
				vport:    "7004",
				toSource: "10.0.3.4",
			},
		},
		kernelRules,
	)
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
