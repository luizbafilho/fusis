package iptables

import (
	"testing"

	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/net"
	"github.com/luizbafilho/fusis/state"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type IptablesSuite struct {
	iptablesMngr *IptablesMngr
}

var _ = Suite(&IptablesSuite{})

func (s *IptablesSuite) SetUpSuite(c *C) {
	var err error
	s.iptablesMngr, err = New(defaultConfig())
	c.Assert(err, IsNil)
}

func (s *IptablesSuite) SetUpTest(c *C) {
}

func (s *IptablesSuite) TearDownSuite(c *C) {
	rules, err := s.iptablesMngr.getSnatRules()
	c.Assert(err, IsNil)

	for _, r := range rules {
		err := s.iptablesMngr.removeRule(r)
		c.Assert(err, IsNil)
	}
}

func (s *IptablesSuite) TestSync(c *C) {
	state, err := state.New(defaultConfig())
	c.Assert(err, IsNil)

	s1 := &types.Service{
		Name:         "test",
		Host:         "10.0.1.1",
		Port:         80,
		Mode:         "nat",
		Protocol:     "tcp",
		Destinations: []types.Destination{},
	}
	state.AddService(s1)

	state.AddService(&types.Service{
		Name:         "test2",
		Host:         "10.0.1.2",
		Port:         80,
		Protocol:     "tcp",
		Mode:         "nat",
		Destinations: []types.Destination{},
	})

	toSource, err := net.GetIpByInterface("eth0")
	c.Assert(err, IsNil)
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

	err = s.iptablesMngr.addRule(rule2)
	c.Assert(err, IsNil)
	err = s.iptablesMngr.addRule(rule3)
	c.Assert(err, IsNil)

	err = s.iptablesMngr.Sync(*state)
	c.Assert(err, IsNil)

	rules, err := s.iptablesMngr.getKernelRulesSet()
	c.Assert(err, IsNil)

	rule1, err := s.iptablesMngr.serviceToSnatRule(*s1)
	c.Assert(err, IsNil)

	c.Assert(rules.Contains(rule2, *rule1), Equals, true)
}

func defaultConfig() *config.BalancerConfig {
	return &config.BalancerConfig{
		PublicInterface:  "eth0",
		PrivateInterface: "eth0",
		Name:             "Test",
		ConfigPath:       "/tmp/test",
		Bootstrap:        true,
		Provider: config.Provider{
			Type: "none",
			Params: map[string]string{
				"interface": "eth0",
				"vip-range": "192.168.0.0/28",
			},
		},
	}
}
