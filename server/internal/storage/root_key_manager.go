package storage

import (
	bolt "go.etcd.io/bbolt"
)

func (s *Storage) GetToken() ([]byte, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	var token []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(metaBucketName)
		if b == nil {
			return ErrFailedToOpenTopBucket
		}

		token = b.Get(metaTokenName)

		return nil
	})

	return token, err
}

func (s *Storage) SetToken(token []byte) error {
	s.m.Lock()
	defer s.m.Unlock()
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(metaBucketName)
		if b == nil {
			return ErrFailedToOpenTopBucket
		}

		return b.Put(metaTokenName, token)
	})
}
