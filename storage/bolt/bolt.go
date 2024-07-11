package bolt

import (
	"fmt"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

type Bolt struct {
	db      *bolt.DB
	rwMutex sync.RWMutex
}

type Record struct {
	Key   []byte
	Value []byte
}

type BucketInfo struct {
	Buckets []string `json:"buckets"`
	KVs     []Record `json:"kvs"`
}

type BucketFullInfo struct {
	Buckets map[string]*BucketFullInfo
	KVS     []Record
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

func openBucketByPath(prefix []string, bucket *bolt.Bucket) (*bolt.Bucket, error) {
	for _, pathPart := range prefix {
		if pathPart == "" {
			return nil, fmt.Errorf("empty bucket name")
		}
		bucket = bucket.Bucket([]byte(pathPart))
		if bucket == nil {
			return nil, fmt.Errorf("failed to open bucket: %s", pathPart)
		}
	}
	return bucket, nil
}

func (b *Bolt) Get(prefix []string, key []byte, bucketName []byte) ([]byte, error) {
	b.rwMutex.RLock()
	defer b.rwMutex.RUnlock()
	var value []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName) // kv
		if b == nil {
			return fmt.Errorf("failed to lookup DB")
		}

		b, err := openBucketByPath(prefix, b)
		if err != nil {
			return err
		}

		value = b.Get(key)
		return nil
	})
	return value, err
}

func (b *Bolt) Set(prefix []string, key, value []byte, bucketName []byte) error {
	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName) // kv
		if b == nil {
			return fmt.Errorf("failed to lookup bolt DB")
		}

		var err error
		for _, pathPart := range prefix {
			b, err = b.CreateBucketIfNotExists([]byte(pathPart))
			if err != nil {
				return err
			}
		}

		if b == nil {
			return fmt.Errorf("incorrect path")
		}

		return b.Put(key, value)
	})
}

func (b *Bolt) Delete(prefix []string, key []byte, bucketName []byte) (int, error) {
	bucketMap := make([]bool, len(prefix)) // true если бакет пуст
	deletedParts := 0

	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()

	err := b.db.Update(func(tx *bolt.Tx) error {
		topLevelBucket := tx.Bucket(bucketName) // kv
		if b == nil {
			return fmt.Errorf("failed to lookup bolt DB")
		}

		var b *bolt.Bucket

		if prefix[0] != "" {
			b = topLevelBucket.Bucket([]byte(prefix[0]))
			if b == nil {
				return fmt.Errorf("the user doesn't have any buckets")
			}
		} else if prefix != nil { // единственный некорректный случай
			return fmt.Errorf("empty username and non-empty prefix")
		}

		for i, pathPart := range prefix[1:] {
			i++
			b = b.Bucket([]byte(pathPart))
			if b == nil {
				return fmt.Errorf("incorrect path")
			}

			c := b.Cursor()
			k, _ := c.First()
			if i+1 < len(prefix) && string(k) == prefix[i+1] {
				k, _ = c.Next()
			}
			bucketMap[i] = k == nil
		}

		err := b.Delete(key)
		if err != nil {
			return err
		}

		if len(prefix) == 1 {
			return nil
		}

		k, _ := b.Cursor().First()
		bucketMap[len(prefix)-1] = k == nil

		// вторая часть транзакции: удаление пустых бакетов

		b = topLevelBucket.Bucket([]byte(prefix[0]))

		indexToDelete := -1
		isAllPathEmpty := true
		for i := len(prefix) - 1; i >= 1; i-- {
			isAllPathEmpty = isAllPathEmpty && bucketMap[i]
			if !bucketMap[i] {
				indexToDelete = i + 1
				if indexToDelete == len(prefix) {
					indexToDelete = -1
				}
				break
			}
		}

		if isAllPathEmpty && len(prefix) > 0 {
			deletedParts = len(prefix) - 1
			return b.DeleteBucket([]byte(prefix[1]))
		}

		if indexToDelete == -1 {
			return nil
		}

		for i, pathPart := range prefix[1:] {
			i++
			if i == indexToDelete {
				deletedParts = len(prefix) - i
				err := b.DeleteBucket([]byte(pathPart))
				if err != nil {
					deletedParts = 0
					return fmt.Errorf("failed to delete 'empty' path in boltdb")
				}
				return nil
			}
			b = b.Bucket([]byte(pathPart))
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return deletedParts, nil
}

func (b *Bolt) ListKV(prefix []string) (*BucketInfo, error) {
	valueCh := make(chan Record)
	bucketCh := make(chan string)
	done := make(chan struct{})

	var valList []Record
	var bucketList []string

	go func() {
		defer close(valueCh)
		defer close(bucketCh)
		defer close(done)
		for {
			select {
			case v := <-valueCh:
				valList = append(valList, v)
			case b := <-bucketCh:
				bucketList = append(bucketList, b)
			case <-done:
				return
			}
		}
	}()

	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(kvBucketName)
		if b == nil {
			return fmt.Errorf("failed to lookup bolt DB")
		}

		b, err := openBucketByPath(prefix, b)
		if err != nil {
			return err
		}

		err = b.ForEach(func(k, v []byte) error {
			if v == nil {
				bucketCh <- string(k)
			} else {
				valueCh <- Record{k, v}
			}
			return nil
		})

		done <- struct{}{}

		return err
	})

	return &BucketInfo{bucketList, valList}, err
}

func iterateBucket(b *bolt.Bucket, ch chan Record, done chan struct{}, kvsCh chan []Record) (*BucketFullInfo, error) {
	if b == nil {
		return nil, nil
	}

	bucketList := make([][]byte, 0)
	curBucket := &BucketFullInfo{}

	err := b.ForEach(func(k, v []byte) error {
		if v == nil {
			bucketList = append(bucketList, k)
		} else {
			ch <- Record{k, v}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	done <- struct{}{}
	curBucket.KVS = <-kvsCh
	curBucket.Buckets = make(map[string]*BucketFullInfo, len(bucketList))
	for _, bucketName := range bucketList {
		curBucket.Buckets[string(bucketName)], err = iterateBucket(b.Bucket(bucketName), ch, done, kvsCh)
		if err != nil {
			return nil, err
		}
	}
	return curBucket, nil
}

func (b *Bolt) ShowBucketRecursion(prefix []string, bucketName []byte) (*BucketFullInfo, error) {
	kvCh := make(chan Record)
	kvsCh := make(chan []Record)
	arrCh := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer close(kvCh)
		defer close(done)
		defer close(kvsCh)
		kvs := make([]Record, 0)
		for {
			select {
			case r := <-kvCh:
				kvs = append(kvs, r)
			case <-arrCh:
				kvsCh <- kvs
				kvs = make([]Record, 0)
			case <-done:
				return
			}
		}
	}()

	BucketInfo := &BucketFullInfo{}
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(kvBucketName)
		if b == nil {
			return fmt.Errorf("failed to lookup bolt DB")
		}

		b, err := openBucketByPath(prefix, b)
		if err != nil {
			return err
		}

		BucketInfo, err = iterateBucket(b, kvCh, arrCh, kvsCh)

		return err
	})

	return BucketInfo, err
}
