package fusis

import (
	"encoding/json"
	"fmt"

	"github.com/luizbafilho/fusis/api/types"
	"github.com/luizbafilho/fusis/engine"
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
	return b.engine.State.GetServices()
}

// AddService ...
func (b *Balancer) AddService(svc *types.Service) error {
	b.Lock()
	defer b.Unlock()

	for _, existingSvc := range b.engine.State.GetServices() {
		if existingSvc.GetId() == svc.GetId() {
			return types.ErrServiceAlreadyExists
		}
	}

	if err := b.provider.AllocateVIP(svc, b.engine.State); err != nil {
		return err
	}

	c := &engine.Command{
		Op:      engine.AddServiceOp,
		Service: svc,
	}

	if err := b.ApplyToRaft(c); err != nil {
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
	return b.engine.State.GetService(name)
}

func (b *Balancer) DeleteService(name string) error {
	b.Lock()
	defer b.Unlock()

	svc, err := b.engine.State.GetService(name)
	if err != nil {
		return err
	}

	c := &engine.Command{
		Op:      engine.DelServiceOp,
		Service: svc,
	}

	return b.ApplyToRaft(c)
}

func (b *Balancer) GetDestination(name string) (*types.Destination, error) {
	b.Lock()
	defer b.Unlock()
	return b.engine.State.GetDestination(name)
}

func (b *Balancer) AddDestination(svc *types.Service, dst *types.Destination) error {
	b.Lock()
	defer b.Unlock()
	c := &engine.Command{
		Op:          engine.AddDestinationOp,
		Service:     svc,
		Destination: dst,
	}

	return b.ApplyToRaft(c)
}

func (b *Balancer) DeleteDestination(dst *types.Destination) error {
	b.Lock()
	defer b.Unlock()
	svc, err := b.engine.State.GetService(dst.ServiceId)
	if err != nil {
		return err
	}

	c := &engine.Command{
		Op:          engine.DelDestinationOp,
		Service:     svc,
		Destination: dst,
	}

	return b.ApplyToRaft(c)
}

func (b *Balancer) ApplyToRaft(cmd *engine.Command) error {
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
