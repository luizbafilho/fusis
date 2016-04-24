package ipvs

import (
	"github.com/asdine/storm"
	log "github.com/golang/glog"
	"github.com/pborman/uuid"
)

func initStore(s *storm.DB) error {
	if err := s.Init(&Service{}); err != nil {
		log.Errorf("Service bucket creation failed: %v", err)
		return err
	}

	if err := s.Init(&Destination{}); err != nil {
		log.Errorf("Destination bucket creation failed: %v", err)
		return err
	}

	Store = &IpvsStore{s}

	return nil
}

func (s *IpvsStore) AddService(svc *Service) error {
	// svc.Id = uuid.New()
	err := Store.db.Save(svc)
	if err != nil {
		log.Errorf("store.AddService failed: %v", err)
		return err
	}
	return nil
}

func (s *IpvsStore) GetServices() (*[]Service, error) {
	svcs := []Service{}

	err := Store.db.All(&svcs)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(svcs); i++ {
		err := s.getDestinations(&svcs[i])
		if err != nil {
			return nil, err
		}
	}

	return &svcs, nil
}

func (s *IpvsStore) GetService(name string) (*Service, error) {
	var svc Service
	err := Store.db.One("Name", name, &svc)
	if err != nil {
		return nil, err
	}

	if err := s.getDestinations(&svc); err != nil {
		return nil, err
	}

	return &svc, nil
}

func (s *IpvsStore) DeleteService(svc *Service) error {
	var dsts []Destination
	Store.db.Find("ServiceId", svc.GetId(), &dsts)

	for _, d := range dsts {
		if err := Store.db.Remove(d); err != nil {
			return err
		}
	}

	return Store.db.Remove(svc)
}

func (s *IpvsStore) AddDestination(dst *Destination) error {
	dst.Id = uuid.New()
	err := Store.db.Save(dst)
	if err != nil {
		log.Errorf("store.AddDestination failed: %v", err)
		return err
	}
	return nil
}

func (s *IpvsStore) GetDestination(name string) (*Destination, error) {
	var dst Destination

	err := Store.db.One("Name", name, &dst)
	if err != nil {
		return nil, err
	}

	return &dst, nil
}

func (s *IpvsStore) DeleteDestination(dst *Destination) error {
	return Store.db.Remove(dst)
}

func (s *IpvsStore) getDestinations(svc *Service) error {
	dsts := []Destination{}
	err := Store.db.Find("ServiceId", svc.GetId(), &dsts)
	if err != nil && err != storm.ErrNotFound {
		return err
	}
	svc.Destinations = dsts
	return nil
}
