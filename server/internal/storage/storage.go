package storage

import (
	"sync"
	"time"

	"github.com/liriquew/secret_storage/server/internal/lib/config"
	bolt "go.etcd.io/bbolt"
)

type Storage struct {
	db *bolt.DB
	m  sync.RWMutex
}

func New(cfg config.StorageConfig) (*Storage, error) {
	db, err := bolt.Open(cfg.Path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(recordsBucketName)
		if err == nil {
			_, err = tx.CreateBucketIfNotExists(userBucketName)
		}
		if err == nil {
			_, err = tx.CreateBucketIfNotExists(metaBucketName)
		}
		return err
	})

	if err != nil {
		return nil, err
	}

	return &Storage{db, sync.RWMutex{}}, nil
}

func (s *Storage) Close() error {
	s.m.Lock()
	defer s.m.Unlock()
	return s.db.Close()
}
