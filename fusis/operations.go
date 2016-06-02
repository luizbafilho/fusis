package fusis

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/engine"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/pborman/uuid"
)

// GetServices get all services
func (b *Balancer) GetServices() *[]ipvs.Service {
	return b.engine.State.GetServices()
}

// AddService ...
func (b *Balancer) AddService(svc *ipvs.Service) error {
	b.Lock()
	defer b.Unlock()

	if err := b.engine.Provider.AllocateVIP(svc); err != nil {
		return err
	}

	svc.Id = uuid.New()

	c := &engine.Command{
		Op:      engine.AddServiceOp,
		Service: svc,
	}

	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := b.raft.Apply(bytes, raftTimeout)
	if err, ok := f.(error); ok {
		if err := b.engine.Provider.ReleaseVIP(*svc); err != nil {
			return err
		}

		return err
	}

	return nil
}

//GetService get a service
func (b *Balancer) GetService(name string) (*ipvs.Service, error) {
	return b.engine.State.GetService(name)
}

func (b *Balancer) DeleteService(name string) error {
	log.Infof("Deleting Service: %v", name)

	svc, err := b.GetService(name)
	if err != nil {
		return err
	}

	c := &engine.Command{
		Op:      engine.DelServiceOp,
		Service: svc,
	}

	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := b.raft.Apply(bytes, raftTimeout)
	if err, ok := f.(error); ok {
		return err
	}

	return nil
}

func (b *Balancer) GetDestination(name string) (*ipvs.Destination, error) {
	return b.engine.State.GetDestination(name)
}

func (b *Balancer) AddDestination(svc *ipvs.Service, dst *ipvs.Destination) error {
	dst.Id = uuid.New()

	c := &engine.Command{
		Op:          engine.AddDestinationOp,
		Service:     svc,
		Destination: dst,
	}

	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := b.raft.Apply(bytes, raftTimeout)
	if err, ok := f.(error); ok {
		return err
	}

	return nil
}

func (b *Balancer) DeleteDestination(dst *ipvs.Destination) error {
	svc, err := b.GetService(dst.ServiceId)
	if err != nil {
		return err
	}

	c := &engine.Command{
		Op:          engine.DelDestinationOp,
		Service:     svc,
		Destination: dst,
	}

	bytes, err := json.Marshal(c)
	if err != nil {
		return err
	}

	f := b.raft.Apply(bytes, raftTimeout)
	if err, ok := f.(error); ok {
		return err
	}

	return nil
}
