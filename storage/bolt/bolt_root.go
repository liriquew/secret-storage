package bolt

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

func (b *Bolt) GetToken() ([]byte, error) {
	b.rwMutex.RLock()
	defer b.rwMutex.RUnlock()
	var token []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(metaBucketName)
		if b == nil {
			return fmt.Errorf("bucket %q not found", metaBucketName)
		}
		t := b.Get(metaTokenName)
		if t != nil {
			token = make([]byte, len(t))
			copy(token, t)
		}
		return nil
	})
	return token, err
}

func (b *Bolt) SetToken(token []byte) error {
	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(metaBucketName)
		return b.Put(metaTokenName, token)
	})
}
