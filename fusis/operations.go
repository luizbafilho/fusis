package fusis

import (
	"fmt"

	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/health"
)

const (
	ConfigInterfaceAgentQuery = "config-interface-agent"
)

type ErrCrashError struct {
	original error
}

func (e ErrCrashError) Error() string {
	return fmt.Sprintf("unable to apply commited log, inconsistent routing state, leaving cluster. original error: %s", e.original)
}

// GetServices get all services
func (b *Balancer) GetServices() []types.Service {
	b.Lock()
	defer b.Unlock()
	return b.state.GetServices()
}

func (b *Balancer) GetDestinations(svc *types.Service) []types.Destination {
	b.Lock()
	defer b.Unlock()

	return b.state.GetDestinations(svc)
}

// AddService ...
func (b *Balancer) AddService(svc *types.Service) error {
	return nil
}

//GetService get a service
func (b *Balancer) GetService(name string) (*types.Service, error) {
	b.Lock()
	defer b.Unlock()
	return b.state.GetService(name)
}

func (b *Balancer) DeleteService(name string) error {
	return nil
}

func (b *Balancer) GetDestination(name string) (*types.Destination, error) {
	b.Lock()
	defer b.Unlock()
	return b.state.GetDestination(name)
}

func (b *Balancer) AddDestination(svc *types.Service, dst *types.Destination) error {
	return nil
}

type AgentInterfaceConfig struct {
	ServiceAddress string
	Mode           string
}

func (b *Balancer) setupDestination(svc *types.Service, dst *types.Destination) error {
	// params := serf.QueryParam{
	// 	FilterNodes: []string{dst.Name},
	// }
	//
	// config := AgentInterfaceConfig{
	// 	ServiceAddress: svc.Address,
	// 	Mode:           svc.Mode,
	// }
	//
	// payload, _ := json.Marshal(config)
	// _, err := b.serf.Query(ConfigInterfaceAgentQuery, payload, &params)
	// if err != nil {
	// 	logrus.Errorf("Balancer: add-balancer event error: %v", err)
	// 	return err
	// }
	return nil
}

func (b *Balancer) DeleteDestination(dst *types.Destination) error {
	return nil
}

func (b *Balancer) AddCheck(dst *types.Destination) error {
	return nil
}

func (b *Balancer) DelCheck(dst *types.Destination) error {
	return nil
}

func (b *Balancer) UpdateCheck(check health.Check) error {
	return nil
}
