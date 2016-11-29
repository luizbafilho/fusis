package fusis

import (
	"errors"
	"time"

	"github.com/luizbafilho/fusis/types"
)

const (
	ConfigInterfaceAgentQuery = "config-interface-agent"
)

var (
	ErrResourceNotFound = errors.New("Resource not found")
	ErrResourceConflict = errors.New("Resource conflict")
)

// GetServices get all services
func (b *FusisBalancer) GetServices() []types.Service {
	return b.state.GetServices()
}

func (b *FusisBalancer) GetDestinations(svc *types.Service) []types.Destination {
	return b.state.GetDestinations(svc)
}

// AddService ...
func (b *FusisBalancer) AddService(svc *types.Service) error {
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
func (b *FusisBalancer) GetService(name string) (*types.Service, error) {
	return b.state.GetService(name)
}

func (b *FusisBalancer) DeleteService(name string) error {
	svc, err := b.state.GetService(name)
	if err != nil {
		return err
	}

	return b.store.DeleteService(svc)
}

func (b *FusisBalancer) GetDestination(name string) (*types.Destination, error) {
	return b.state.GetDestination(name)
}

func (b *FusisBalancer) AddDestination(svc *types.Service, dst *types.Destination) error {
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

func (b *FusisBalancer) setupDestination(svc *types.Service, dst *types.Destination) error {
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
	// 	logrus.Errorf("FusisBalancer: add-balancer event error: %v", err)
	// 	return err
	// }
	return nil
}

func (b *FusisBalancer) DeleteDestination(dst *types.Destination) error {
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

func (b *FusisBalancer) AddCheck(check types.CheckSpec) error {
	if check.Timeout == 0 {
		check.Timeout = 5 * time.Second
	}

	if check.Interval == 0 {
		check.Interval = 10 * time.Second
	}

	return b.store.AddCheck(check)
}

func (b *FusisBalancer) DeleteCheck(check types.CheckSpec) error {
	return b.store.DeleteCheck(check)
}
