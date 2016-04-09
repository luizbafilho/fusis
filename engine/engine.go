package engine

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/infra"
)

var store *BoltDB
var cs *infra.CloudstackIaaS

func init() {
	var err error
	store, err = NewStore()
	if err != nil {
		panic(err)
	}
	cs = infra.NewCloudstackIaaS("0b5b922f-6b71-4955-b6bf-250685323dc9", "vr5P_5mC_H7vN1MDRQqotbW8h6EEjjnIGrDiqhLEyHJHY8lb_wznIDkeNPgjfmv45M4PCqkRX6fzxk5bMY_etQ", "rz7-Hek8YpblTb8wOXj-oaK6ZW2sAIF_Ph7Wy53q2GLLWNrAe1px3LAGW23OW3KanOUz1OHEatLOJb1WDK8Cvw")
}

func GetServices() ([]*Service, error) {
	return store.GetServices()
}

func AddService(svc *Service) error {
	service, _ := store.GetService(svc.GetId())
	if service != nil {
		return fmt.Errorf("Service already exists: %+v", service)
	}

	// ip, err := cs.SetVip("fusis")
	// if err != nil {
	// 	return err
	// }

	// ip := "10.2.3.4"
	// svc.Host = ip

	if err := store.AddService(svc); err != nil {
		return err
	}

	if err := IPVSAddService(svc.ToIpvsService()); err != nil {
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
	if err := IPVSAddDestination(*svc.ToIpvsService(), *dst.ToIpvsDestination()); err != nil {
		return err
	}

	if err := store.AddDestination(dst); err != nil {
		return err
	}

	service, err := store.GetService(dst.ServiceId)
	if err != nil {
		return err
	}

	dsts := service.Destinations
	dsts = append(dsts, *dst)
	service.Destinations = dsts

	err = store.AddService(service)
	if err != nil {
		return err
	}

	return nil
}

func DeleteDestination(id string) error {
	log.Infof("Deleting Destination: %v", id)
	dst := store.GetDestination(id)

	svc, err := GetService(dst.ServiceId)
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
