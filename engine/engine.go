package engine

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
)

var store *BoltDB

func init() {
	var err error
	store, err = NewStore()
	if err != nil {
		fmt.Println("=======> falhando aqui")
		panic(err)
	}
}

func GetServices() (*[]Service, error) {
	ipvsSvcs, err := IPVSGetServices()
	if err != nil {
		return nil, err
	}

	services := []Service{}
	for _, svc := range ipvsSvcs {
		services = append(services, NewService(svc))
	}

	return &services, nil
}

func AddService(svc *Service) error {
	if err := IPVSAddService(svc.ToIpvsService()); err != nil {
		return err
	}

	if err := store.AddService(svc); err != nil {
		return err
	}

	return nil
}

func AddDestination(svc *Service, dst *Destination) error {
	log.Infof("Adding Destination: %v", dst.Name)
	if err := IPVSAddDestination(*svc.ToIpvsService(), *dst.ToIpvsDestination()); err != nil {
		return err
	}

	if err := store.AddDestination(dst); err != nil {
		return err
	}

	return nil
}

func DeleteDestination(id string) error {
	log.Infof("Deleting Destination: %v", id)
	dst := store.GetDestination(id)

	svc, err := GetServiceFromId(dst.ServiceId)
	if err != nil {
		panic(err)
	}

	if err := IPVSDeleteDestination(*svc.ToIpvsService(), *dst.ToIpvsDestination()); err != nil {
		return err
	}

	if err := store.DeleteDestination(dst); err != nil {
		return err
	}

	return nil
}
