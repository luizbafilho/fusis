package engine

import (
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
	s.Lock()
	defer s.Unlock()

	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("services"))

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
		fmt.Println("=========> erro no update")
		return err
	}
	return nil
}

func (s *BoltDB) GetService(serviceId string) (*Service, error) {
	return nil, nil
}
