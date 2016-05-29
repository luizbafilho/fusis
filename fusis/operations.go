package fusis

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/engine"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/pborman/uuid"
)

// GetServices get all services
func GetServices() (*[]ipvs.Service, error) {
	return ipvs.Store.GetServices()
}

// AddService ...
func (b *Balancer) AddService(svc *ipvs.Service) error {
	b.Lock()
	defer b.Unlock()

	if err := b.provider.AllocateVIP(svc); err != nil {
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
		if err := b.provider.ReleaseVIP(*svc); err != nil {
			return err
		}

		return err
	}

	return nil
}

//GetService get a service
func GetService(id string) (*ipvs.Service, error) {
	return ipvs.Store.GetService(id)
}

func (b *Balancer) DeleteService(id string) error {
	log.Infof("Deleting Service: %v", id)

	svc, err := ipvs.Store.GetService(id)
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

func GetDestination(name string) (*ipvs.Destination, error) {
	return ipvs.Store.GetDestination(name)
}

func (b *Balancer) AddDestination(svc *ipvs.Service, dst *ipvs.Destination) error {
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

func DeleteDestination(id string) error {
	log.Infof("Deleting Destination: %v", id)
	dst, err := ipvs.Store.GetDestination(id)
	if err != nil {
		return err
	}

	svc, err := ipvs.Store.GetService(dst.ServiceId)
	if err != nil {
		return err
	}

	if err := ipvs.Store.DeleteDestination(dst); err != nil {
		return err
	}
	if err := ipvs.Kernel.DeleteDestination(*svc.ToIpvsService(), *dst.ToIpvsDestination()); err != nil {
		return err
	}

	return nil
}
