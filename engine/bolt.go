package engine

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	log "github.com/golang/glog"

	"github.com/boltdb/bolt"
)

type BoltDB struct {
	db *bolt.DB
	sync.Mutex
}

var (
	services     = []byte("services")
	destinations = []byte("destinations")
)

func NewStore() (*BoltDB, error) {
	db, err := bolt.Open("fusis.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Errorf("Bolt opening file failed: %v", err)
		return nil, err
	}
	// defer db.Close()

	if err = createBucket(db, "services"); err != nil {
		log.Errorf("Service bucket creation failed: %v", err)
		return nil, err
	}

	if err = createBucket(db, "destinations"); err != nil {
		log.Errorf("Destination bucket creation failed: %v", err)
		return nil, err
	}

	return &BoltDB{
		db: db,
	}, nil
}

func createBucket(db *bolt.DB, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

func (s *BoltDB) AddService(svc *Service) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(services)

		json, err := svc.ToJson()
		if err != nil {
			return err
		}

		err = b.Put([]byte(svc.GetId()), json)
		if err != nil {
			log.Errorf("store.AddService failed: %v", err)
		}
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *BoltDB) GetService(id string) (*Service, error) {
	var svc Service

	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(services)
		v := b.Get([]byte(id))

		if v == nil {
			return fmt.Errorf("Services not found: %+v", id)
		}

		if err := json.Unmarshal(v, &svc); err != nil {
			return nil
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &svc, nil
}

func (s *BoltDB) DeleteService(svc *Service) error {
	if err := s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(services).Delete([]byte(svc.GetId()))
	}); err != nil {
		return err
	}

	return nil
}

func (s *BoltDB) AddDestination(dst *Destination) error {
	if err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(destinations)

		json, err := dst.ToJson()
		if err != nil {
			return err
		}

		err = b.Put([]byte(dst.GetId()), json)
		if err != nil {
			log.Errorf("store.AddDestination failed: %v", err)
		}
		return err
	}); err != nil {
		return err
	}

	return nil
}

func (s *BoltDB) GetDestination(id string) *Destination {
	var dst Destination

	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(destinations)
		v := b.Get([]byte(id))
		if err := json.Unmarshal(v, &dst); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil
	}

	return &dst
}

func (s *BoltDB) DeleteDestination(dst *Destination) error {
	if err := s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(destinations).Delete([]byte(dst.Name))
	}); err != nil {
		return err
	}

	return nil
}
