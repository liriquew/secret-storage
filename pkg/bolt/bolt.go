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

type Record struct {
	Key   []byte
	Value []byte
}

var (
	kvBucketName   = []byte("kv")
	userBucketName = []byte("user")
	metaBucketName = []byte("meta")
	metaTokenName  = []byte("token")
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
		if err == nil {
			_, err = tx.CreateBucketIfNotExists(metaBucketName)
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

func (b *Bolt) Get(key, bucketName []byte) ([]byte, error) {
	b.rwMutex.RLock()
	defer b.rwMutex.RUnlock()
	var value []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName) // kv
		if b == nil {
			return fmt.Errorf("failed to lookup DB")
		}
		value = b.Get(key)
		return nil
	})
	return value, err
}

func (b *Bolt) Set(key, value, bucketName []byte) error {
	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName) // kv
		if b == nil {
			return fmt.Errorf("failed to lookup bolt DB")
		}
		return b.Put(key, value)
	})
}

func (b *Bolt) Delete(key, bucketName []byte) error {
	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName) // kv
		if b == nil {
			return fmt.Errorf("failed to lookup bolt DB")
		}
		return b.Delete(key)
	})
}

func (b *Bolt) ListKV() ([]Record, error) {
	var list []Record
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(kvBucketName)
		if b == nil {
			return fmt.Errorf("failed to lookup bolt DB")
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			list = append(list, Record{k, v})
		}
		return nil
	})
	return list, err
}

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

func (b *Bolt) ShowList() ([]Record, error) {
	fmt.Println("------KV------")
	var list []Record
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(kvBucketName)
		if b == nil {
			return fmt.Errorf("failed to lookup bolt DB")
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
			list = append(list, Record{k, v})
		}
		return nil
	})

	if err != nil {
		return nil, err
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
	return list, err
}
