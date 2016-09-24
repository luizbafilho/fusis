package fusis

import (
	"encoding/json"
	"fmt"

	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/health"
	"github.com/luizbafilho/fusis/state"
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
	b.Lock()
	defer b.Unlock()

	_, err := b.state.GetService(svc.GetId())
	if err == nil {
		return types.ErrServiceAlreadyExists
	} else if err != types.ErrServiceNotFound {
		return err
	}

	if err = b.ipam.AllocateVIP(svc); err != nil {
		return err
	}

	c := &state.Command{
		Op:      state.AddServiceOp,
		Service: svc,
	}

	if err = b.ApplyToRaft(c); err != nil {
		if e := b.ipam.ReleaseVIP(*svc); e != nil {
			return e
		}
		return err
	}

	return nil
}

//GetService get a service
func (b *Balancer) GetService(name string) (*types.Service, error) {
	b.Lock()
	defer b.Unlock()
	return b.state.GetService(name)
}

func (b *Balancer) DeleteService(name string) error {
	b.Lock()
	defer b.Unlock()

	svc, err := b.state.GetService(name)
	if err != nil {
		return err
	}

	c := &state.Command{
		Op:      state.DelServiceOp,
		Service: svc,
	}

	return b.ApplyToRaft(c)
}

func (b *Balancer) GetDestination(name string) (*types.Destination, error) {
	b.Lock()
	defer b.Unlock()
	return b.state.GetDestination(name)
}

func (b *Balancer) AddDestination(svc *types.Service, dst *types.Destination) error {
	b.Lock()
	defer b.Unlock()

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

	c := &state.Command{
		Op:          state.AddDestinationOp,
		Service:     svc,
		Destination: dst,
	}

	return b.ApplyToRaft(c)
}

func (b *Balancer) DeleteDestination(dst *types.Destination) error {
	b.Lock()
	defer b.Unlock()
	svc, err := b.state.GetService(dst.ServiceId)
	if err != nil {
		return err
	}

	_, err = b.state.GetDestination(dst.GetId())
	if err != nil {
		return err
	}

	c := &state.Command{
		Op:          state.DelDestinationOp,
		Service:     svc,
		Destination: dst,
	}

	return b.ApplyToRaft(c)
}

func (b *Balancer) AddCheck(dst *types.Destination) error {
	c := &state.Command{
		Op:          state.AddCheckOp,
		Destination: dst,
	}

	return b.ApplyToRaft(c)
}

func (b *Balancer) DelCheck(dst *types.Destination) error {
	c := &state.Command{
		Op:          state.DelCheckOp,
		Destination: dst,
	}

	return b.ApplyToRaft(c)
}

func (b *Balancer) UpdateCheck(check health.Check) error {
	c := &state.Command{
		Op:    state.UpdateCheckOp,
		Check: &check,
	}

	return b.ApplyToRaft(c)
}

func (b *Balancer) ApplyToRaft(cmd *state.Command) error {
	bytes, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	f := b.raft.Apply(bytes, raftTimeout)
	if err = f.Error(); err != nil {
		return err
	}
	rsp := f.Response()
	if err, ok := rsp.(error); ok {
		return ErrCrashError{original: err}
	}
	return nil
}
