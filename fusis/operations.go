package fusis

import (
	"encoding/json"
	"fmt"

	"github.com/luizbafilho/fusis/api/types"
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
	return b.state.Store.GetServices()
}

// AddService ...
func (b *Balancer) AddService(svc *types.Service) error {
	b.Lock()
	defer b.Unlock()

	_, err := b.state.Store.GetService(svc.GetId())
	if err == nil {
		return types.ErrServiceAlreadyExists
	} else if err != types.ErrServiceNotFound {
		return err
	}

	if err = b.provider.AllocateVIP(svc, b.state.Store); err != nil {
		return err
	}

	c := &state.Command{
		Op:      state.AddServiceOp,
		Service: svc,
	}

	if err = b.ApplyToRaft(c); err != nil {
		if e := b.provider.ReleaseVIP(*svc); e != nil {
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
	return b.state.Store.GetService(name)
}

func (b *Balancer) DeleteService(name string) error {
	b.Lock()
	defer b.Unlock()

	svc, err := b.state.Store.GetService(name)
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
	return b.state.Store.GetDestination(name)
}

func (b *Balancer) AddDestination(svc *types.Service, dst *types.Destination) error {
	b.Lock()
	defer b.Unlock()

	stateSvc, err := b.state.Store.GetService(svc.GetId())
	if err != nil {
		return err
	}

	_, err = b.state.Store.GetDestination(dst.GetId())
	if err == nil {
		return types.ErrDestinationAlreadyExists
	} else if err != types.ErrDestinationNotFound {
		return err
	}

	for _, existDst := range stateSvc.Destinations {
		if existDst.Host == dst.Host && existDst.Port == dst.Port {
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
	svc, err := b.state.Store.GetService(dst.ServiceId)
	if err != nil {
		return err
	}

	_, err = b.state.Store.GetDestination(dst.GetId())
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
