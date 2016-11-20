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
	_, err := b.state.GetService(svc.GetId())
	if err == nil {
		return types.ErrServiceAlreadyExists
	} else if err != types.ErrServiceNotFound {
		return err
	}

	if err = b.ipam.AllocateVIP(svc); err != nil {
		return err
	}

	return b.store.AddService(svc)
}

//GetService get a service
func (b *Balancer) GetService(name string) (*types.Service, error) {
	return b.state.GetService(name)
}

func (b *Balancer) DeleteService(name string) error {
	svc, err := b.state.GetService(name)
	if err != nil {
		return err
	}

	return b.store.DeleteService(svc)
}

func (b *Balancer) GetDestination(name string) (*types.Destination, error) {
	b.Lock()
	defer b.Unlock()
	return b.state.GetDestination(name)
}

func (b *Balancer) AddDestination(svc *types.Service, dst *types.Destination) error {
	_, err := b.state.GetDestination(dst.GetId())
	if err == nil {
		return types.ErrDestinationAlreadyExists
	} else if err != types.ErrDestinationNotFound {
		return err
	}

	for _, existDst := range b.state.GetDestinations(svc) {
		if existDst.Address == dst.Address && existDst.Port == dst.Port {
			return types.ErrDestinationAlreadyExists
		}
	}

	//TODO: Configurate destination
	// if err := b.setupDestination(svc, dst); err != nil {
	// 	return errors.Wrap(err, "setup destination failed")
	// }

	return b.store.AddDestination(svc, dst)
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
	svc, err := b.state.GetService(dst.ServiceId)
	if err != nil {
		return err
	}

	_, err = b.state.GetDestination(dst.GetId())
	if err != nil {
		return err
	}

	return b.store.DeleteDestination(svc, dst)
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
