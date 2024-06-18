package bolt

import (
	"fmt"
	bolt "go.etcd.io/bbolt"
	"sync"
	"time"
)

type Bolt struct {
	db      *bolt.DB
	rwMutex sync.RWMutex
}

var bucketName = []byte("kv")

func New(path string) (*Bolt, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("kv"))
		return err
	})

	if err != nil {
		return nil, err
	}

	return &Bolt{db, sync.RWMutex{}}, nil
}

func (b *Bolt) Close() error {
	return b.db.Close()
}

func (b *Bolt) Put(key, value []byte) error {
	fmt.Println("PUT START")
	b.rwMutex.Lock()
	err := b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		err := b.Put(key, value)
		time.Sleep(time.Second * 15)
		fmt.Println("PUT END")
		return err
	})
	b.rwMutex.Unlock()
	return err
}

func (b *Bolt) Get(key []byte) ([]byte, error) {
	fmt.Println("GET START")
	b.rwMutex.RLock()
	var value []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		value = b.Get(key)
		return nil
	})
	b.rwMutex.RUnlock()
	fmt.Println("GET END")
	return value, err
}

func (b *Bolt) Delete(key []byte) error {
	fmt.Println("DELETE START")
	b.rwMutex.Lock()
	err := b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		return b.Delete(key)
	})
	b.rwMutex.Unlock()
	return err
}

func (b *Bolt) List() ([][]byte, error) {
	var values [][]byte
	fmt.Println("START")
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
		}
		return nil
	})
	fmt.Println("END")
	return values, err
}
