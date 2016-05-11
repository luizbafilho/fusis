package ipvs

import (
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/boltdb/bolt"
	"github.com/pborman/uuid"
)

func InitStore() error {
	log.Info("Initializing IPVS Store")
	boltPath := "fusis.db"

	// Cleaning up any previous state in case of a failure.
	// Raft will be responsible to recover the proper state.
	if err := deleteBoltFile(boltPath); err != nil {
		return err
	}

	db, err := newStormDb(boltPath)
	if err != nil {
		return err
	}

	if err := db.Init(&Service{}); err != nil {
		log.Errorf("Service bucket creation failed: %v", err)
		return err
	}

	if err := db.Init(&Destination{}); err != nil {
		log.Errorf("Destination bucket creation failed: %v", err)
		return err
	}

	Store = &IpvsStore{db}

	return nil
}

// AddService add a new service to bolt
func (s *IpvsStore) AddService(svc *Service) error {
	err := Store.db.Save(svc)
	if err != nil {
		log.Errorf("store.AddService failed: %v", err)
		return err
	}
	return nil
}

// GetServices get all services stored
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

func newStormDb(path string) (*storm.DB, error) {
	db, err := storm.OpenWithOptions(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Errorf("Bolt opening file failed: %v", err)
		return nil, err
	}

	return db, nil
}

func deleteBoltFile(name string) error {
	if fileExists(name) {
		return os.Remove(name)
	}
	return nil
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}
