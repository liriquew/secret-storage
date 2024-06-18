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

var (
	kvBucketName   = []byte("kv")
	userBucketName = []byte("user")
)

func New(path string) (*Bolt, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(kvBucketName)
		if err == nil {
			_, err = tx.CreateBucketIfNotExists(userBucketName)
		}
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
	//fmt.Println("PUT START")
	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(kvBucketName)
		err := b.Put(key, value)
		//fmt.Println("PUT END")
		return err
	})
}

func (b *Bolt) Get(key []byte) ([]byte, error) {
	//fmt.Println("GET START")
	b.rwMutex.RLock()
	var value []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(kvBucketName)
		value = b.Get(key)
		return nil
	})
	b.rwMutex.RUnlock()
	//fmt.Println("GET END")
	return value, err
}

func (b *Bolt) Delete(key []byte) error {
	fmt.Println("DELETE START")
	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(kvBucketName)
		return b.Delete(key)
	})
}

func (b *Bolt) List() error {
	fmt.Println("------KV------")
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(kvBucketName)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
		}
		return nil
	})

	if err != nil {
		return err
	}
	fmt.Println("-----USERS-----")
	err = b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(userBucketName)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
		}
		return nil
	})
	return err
}
