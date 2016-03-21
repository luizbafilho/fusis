package engine

import "fmt"

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
	return IPVSAddDestination(*svc.ToIpvsService(), *dst.ToIpvsDestination())
}
