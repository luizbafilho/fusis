package fusis

import (
	"encoding/json"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/hashicorp/serf/serf"
	"github.com/luizbafilho/fusis/types"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/net"
	"github.com/pkg/errors"
)

const (
	TUNNEL_INTERFACE = "tunl0"
)

type Agent struct {
	serf    *serf.Serf
	eventCh chan serf.Event
	config  *config.AgentConfig
}

func NewAgent(config *config.AgentConfig) (*Agent, error) {
	log.Infof("Fusis Agent: Config ==> %+v", config)
	agent := &Agent{
		eventCh: make(chan serf.Event, 64),
		config:  config,
	}

	return agent, nil
}

func (a *Agent) SetupRouteVip(vip string) error {
	if err := disableArpAnnounce(); err != nil {
		return err
	}

	log.Println("setup route vip. ", vip)
	return net.AddIp(vip+"/32", a.config.Interface)
}

func (a *Agent) SetupTunnelVip(vip string) error {
	if err := disableArpAnnounce(); err != nil {
		return err
	}

	if err := setupTunnelInterface(); err != nil {
		return err
	}

	return net.AddIp(vip+"/32", TUNNEL_INTERFACE)
}

func setupTunnelInterface() error {
	if out, err := exec.Command("modprobe", "-va", "ipip").CombinedOutput(); err != nil {
		return errors.Wrapf(err, "Running modprobe ipip failed with message: `%s`", strings.TrimSpace(string(out)))
	}

	if err := net.SetLinkUp(TUNNEL_INTERFACE); err != nil {
		return errors.Wrap(err, "error setting tunnel link up ")
	}

	return nil
}

func disableArpAnnounce() error {
	if err := net.SetSysctl("net.ipv4.conf.all.arp_announce", "2"); err != nil {
		return errors.Wrap(err, "setting net.ipv4.conf.all.arp_announce to 2 failed")
	}

	if err := net.SetSysctl("net.ipv4.conf.all.arp_ignore", "1"); err != nil {
		return errors.Wrap(err, "setting net.ipv4.conf.all.ignore to 1 failed")
	}

	return nil
}

func (a *Agent) Start() error {
	conf := serf.DefaultConfig()
	conf.Init()
	conf.Tags["role"] = "agent"
	conf.Tags["info"] = a.getInfo()

	bindAddr, err := a.config.GetIpByInterface()
	if err != nil {
		log.Fatal(err)
	}

	conf.NodeName = a.config.Name

	conf.MemberlistConfig.BindAddr = bindAddr
	conf.EventCh = a.eventCh

	serf, err := serf.Create(conf)
	if err != nil {
		return err
	}

	a.serf = serf

	go a.handleEvents()

	return nil
}

func (a *Agent) handleEvents() {
	for {
		select {
		case e := <-a.eventCh:
			switch e.EventType() {
			case serf.EventQuery:
				query := e.(*serf.Query)
				a.handleQuery(query)
			default:
				log.Warnf("Balancer: unhandled Serf Event: %#v", e)
			}
		}
	}
}

func (a *Agent) handleQuery(query *serf.Query) {
	payload := query.Payload
	var config AgentInterfaceConfig
	err := json.Unmarshal(payload, &config)
	if err != nil {
		log.Errorf("Balancer: Unable to Unmarshal: %s", payload)
	}

	log.Println("query received. %v", query)

	switch query.Name {
	case ConfigInterfaceAgentQuery:
		a.configInterface(config)
	}

}

func (a *Agent) configInterface(config AgentInterfaceConfig) {
	var err error
	switch config.Mode {
	case types.TUNNEL:
		err = a.SetupTunnelVip(config.ServiceAddress)
	case types.ROUTE:
		err = a.SetupRouteVip(config.ServiceAddress)
	}

	if err != nil {
		log.Errorf("Error configuring network interface. Config: %v. Err: %v", config, err)
	}
}

func (a *Agent) getInfo() string {
	address, err := a.config.GetIpByInterface()
	if err != nil {
		log.Fatal("Unable to get agent host address", err)
	}

	dst := types.Destination{
		Name:      a.config.Name,
		Address:   address,
		Port:      a.config.Port,
		Weight:    a.config.Weight,
		Mode:      a.config.Mode,
		ServiceId: a.config.Service,
	}

	payload, err := json.Marshal(dst)
	if err != nil {
		log.Fatal("Unable to marshal agent info", err)
	}

	return string(payload)
}

func (a *Agent) Join(existing []string, ignoreOld bool) (n int, err error) {
	log.Infof("Fusis Agent: joining: %v ignore: %v", existing, ignoreOld)
	n, err = a.serf.Join(existing, ignoreOld)
	if n > 0 {
		log.Infof("Fusis Agent: joined: %d nodes", n)
	}
	if err != nil {
		log.Warnf("Fusis Agent: error joining: %v", err)
	}
	return
}

func (a *Agent) Shutdown() {
	if err := a.serf.Leave(); err != nil {
		log.Fatalf("Graceful shutdown failed: %s", err)
	}
}
