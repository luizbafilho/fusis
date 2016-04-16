package engine

import (
	"time"

	"github.com/asdine/storm"
	"github.com/boltdb/bolt"
	log "github.com/golang/glog"
	. "github.com/luizbafilho/fusis/engine/store"
	"github.com/pborman/uuid"
)

type StoreBolt struct {
	db   *storm.DB
	path string
}

func NewStore(path string) (*StoreBolt, error) {
	db, err := storm.OpenWithOptions(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Errorf("Bolt opening file failed: %v", err)
		return nil, err
	}

	store = &StoreBolt{
		db:   db,
		path: path,
	}

	if err = store.init(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (s *StoreBolt) init() error {
	if err := s.db.Init(&Service{}); err != nil {
		log.Errorf("Service bucket creation failed: %v", err)
		return err
	}

	if err := s.db.Init(&Destination{}); err != nil {
		log.Errorf("Destination bucket creation failed: %v", err)
		return err
	}
	return nil
}

func (s *StoreBolt) AddService(svc *Service) error {
	svc.Id = uuid.New()
	err := s.db.Save(svc)
	if err != nil {
		log.Errorf("store.AddService failed: %v", err)
		return err
	}
	return nil
}

func (s *StoreBolt) GetServices() (*[]Service, error) {
	svcs := []Service{}

	err := s.db.All(&svcs)
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

func (s *StoreBolt) GetService(name string) (*Service, error) {
	var svc Service
	err := s.db.One("Name", name, &svc)
	if err != nil {
		return nil, err
	}

	if err := s.getDestinations(&svc); err != nil {
		return nil, err
	}

	return &svc, nil
}

func (s *StoreBolt) DeleteService(svc *Service) error {
	return s.db.Remove(svc)
}

func (s *StoreBolt) AddDestination(dst *Destination) error {
	dst.Id = uuid.New()
	err := s.db.Save(dst)
	if err != nil {
		log.Errorf("store.AddDestination failed: %v", err)
		return err
	}
	return nil
}

func (s *StoreBolt) GetDestination(name string) (*Destination, error) {
	var dst Destination

	err := s.db.One("Name", name, &dst)
	if err != nil {
		return nil, err
	}

	return &dst, nil
}

func (s *StoreBolt) DeleteDestination(dst *Destination) error {
	return s.db.Remove(dst)
}

func (s *StoreBolt) getDestinations(svc *Service) error {
	dsts := []Destination{}
	err := s.db.Find("ServiceId", svc.GetId(), &dsts)
	if err != nil && err != storm.ErrNotFound {
		return err
	}
	svc.Destinations = dsts
	return nil
}
