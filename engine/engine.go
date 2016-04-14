package engine

import (
	log "github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/steps"
)

var store *StoreBolt

func init() {
	var err error
	store, err = NewStore("fusis.db")
	if err != nil {
		panic(err)
	}
}

func GetServices() (*[]Service, error) {
	return store.GetServices()
}

func AddService(svc *Service) error {
	seq := steps.NewSequence(
		setVip{svc},
		addServiceStore{svc},
		addServiceIpvs{svc},
	)

	if err := seq.Execute(); err != nil {
		return err
	}

	return nil
}

func GetService(id string) (*Service, error) {
	return store.GetService(id)
}

func DeleteService(id string) error {
	log.Infof("Deleting Service: %v", id)
	svc, err := store.GetService(id)
	if err != nil {
		return err
	}

	if err := IPVSDeleteService(svc.ToIpvsService()); err != nil {
		return err
	}

	if err := store.DeleteService(svc); err != nil {
		return err
	}

	return nil
}

func AddDestination(svc *Service, dst *Destination) error {
	log.Infof("Adding Destination: %v", dst.Name)
	if err := store.AddDestination(dst); err != nil {
		return err
	}

	if err := IPVSAddDestination(*svc.ToIpvsService(), *dst.ToIpvsDestination()); err != nil {
		return err
	}

	return nil
}

func DeleteDestination(id string) error {
	log.Infof("Deleting Destination: %v", id)
	dst, err := store.GetDestination(id)
	if err != nil {
		return err
	}

	svc, err := store.GetService(dst.ServiceId)
	if err != nil {
		return err
	}

	if err := store.DeleteDestination(dst); err != nil {
		return err
	}
	if err := IPVSDeleteDestination(*svc.ToIpvsService(), *dst.ToIpvsDestination()); err != nil {
		return err
	}

	return nil
}
