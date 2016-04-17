package engine

import (
	log "github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/db"
	"github.com/luizbafilho/fusis/ipam"
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/luizbafilho/fusis/steps"
)

func Init() {
	db, err := db.New("fusis.db")
	if err != nil {
		panic(err)
	}

	ipvs.Init(db)
	ipam.Init(db)
}

func GetServices() (*[]ipvs.Service, error) {
	return ipvs.Store.GetServices()
}

func AddService(svc *ipvs.Service) error {
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

func GetService(id string) (*ipvs.Service, error) {
	return ipvs.Store.GetService(id)
}

func DeleteService(id string) error {
	log.Infof("Deleting ipvs.Service: %v", id)
	svc, err := ipvs.Store.GetService(id)
	if err != nil {
		return err
	}

	if err := ipvs.Kernel.DeleteService(svc.ToIpvsService()); err != nil {
		return err
	}

	if err := ipvs.Store.DeleteService(svc); err != nil {
		return err
	}

	return nil
}

func AddDestination(svc *ipvs.Service, dst *ipvs.Destination) error {
	log.Infof("Adding Destination: %v", dst.Name)
	if err := ipvs.Store.AddDestination(dst); err != nil {
		return err
	}

	if err := ipvs.Kernel.AddDestination(*svc.ToIpvsService(), *dst.ToIpvsDestination()); err != nil {
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
