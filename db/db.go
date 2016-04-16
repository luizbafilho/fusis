package db

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/boltdb/bolt"
)

type DB struct {
	DB   *storm.DB
	path string
}

func New(path string) (*DB, error) {
	db, err := storm.OpenWithOptions(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		logrus.Errorf("Bolt opening file failed: %v", err)
		return nil, err
	}

	return &DB{
		DB:   db,
		path: path,
	}, nil
}
