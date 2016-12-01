package fusis

import (
	"time"

	validator "gopkg.in/go-playground/validator.v9"

	"github.com/luizbafilho/fusis/types"
)

const (
	ConfigInterfaceAgentQuery = "config-interface-agent"
)

// GetServices get all services
func (b *FusisBalancer) GetServices() []types.Service {
	return b.state.GetServices()
}

//GetService get a service
func (b *FusisBalancer) GetService(name string) (*types.Service, error) {
	return b.state.GetService(name)
}

// AddService ...
func (b *FusisBalancer) AddService(svc *types.Service) error {
	// Validating service
	if err := b.validateService(svc); err != nil {
		return err
	}

	// Verifies if service is already in the system
	if _, err := b.state.GetService(svc.GetId()); err == nil {
		return types.ErrServiceConflict
	}

	if err := b.ipam.AllocateVIP(svc); err != nil {
		return err
	}

	return b.store.AddService(svc)
}

func (b *FusisBalancer) DeleteService(name string) error {
	svc, err := b.state.GetService(name)
	if err != nil {
		return err
	}

	return b.store.DeleteService(svc)
}

func (b *FusisBalancer) GetDestinations(svc *types.Service) []types.Destination {
	return b.state.GetDestinations(svc)
}

func (b *FusisBalancer) GetDestination(name string) (*types.Destination, error) {
	return b.state.GetDestination(name)
}

func (b *FusisBalancer) AddDestination(svc *types.Service, dst *types.Destination) error {
	// Validating destination
	if err := b.validateDestination(dst); err != nil {
		return err
	}

	// Verifies if destination is already in the system
	if _, err := b.state.GetDestination(dst.GetId()); err == nil {
		return types.ErrDestinationConflict
	}

	// Check if the new destination is unique
	for _, existDst := range b.state.GetDestinations(svc) {
		if dst.Equal(existDst) {
			return types.ErrDestinationConflict
		}
	}

	// Set defaults
	if dst.Weight == 0 {
		dst.Weight = 1
	}

	if dst.Mode == "" {
		dst.Mode = "nat"
	}

	//TODO: Configurate destination
	// if err := b.setupDestination(svc, dst); err != nil {
	// 	return errors.Wrap(err, "setup destination failed")
	// }

	return b.store.AddDestination(svc, dst)
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
	// Setting default values
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

func (b *FusisBalancer) validateService(svc *types.Service) error {
	if err := b.validate.Struct(svc); err != nil {
		errValidation := types.ErrValidation{Type: "service", Errors: make(map[string]string)}
		for _, err := range err.(validator.ValidationErrors) {
			errValidation.Errors[err.Field()] = getValidationMessage(err)
		}
		return errValidation
	}

	return nil
}

func (b *FusisBalancer) validateDestination(dst *types.Destination) error {
	if err := b.validate.Struct(dst); err != nil {
		errValidation := types.ErrValidation{Type: "destination", Errors: make(map[string]string)}
		for _, err := range err.(validator.ValidationErrors) {
			errValidation.Errors[err.Field()] = getValidationMessage(err)
		}
		return errValidation
	}

	return nil
}
