package db

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/boltdb/bolt"
)

func New(path string) (*storm.DB, error) {
	db, err := storm.OpenWithOptions(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		logrus.Errorf("Bolt opening file failed: %v", err)
		return nil, err
	}

	return db, nil
}
