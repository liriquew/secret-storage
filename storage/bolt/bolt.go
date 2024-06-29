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

type RecordOut struct {
	KV     Record
	Indent string
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

func openBucketByPath(prefix []string, bucket *bolt.Bucket) *bolt.Bucket {
	for _, pathPart := range prefix {
		fmt.Println("CURR:", pathPart)
		bucket = bucket.Bucket([]byte(pathPart))
		if bucket == nil {
			return nil
		}
	}
	return bucket
}

func (b *Bolt) Get(username string, prefix []string, key []byte, bucketName []byte) ([]byte, error) {
	b.rwMutex.RLock()
	defer b.rwMutex.RUnlock()
	var value []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName) // kv
		if b == nil {
			return fmt.Errorf("failed to lookup DB")
		}

		if username != "" {
			b = b.Bucket([]byte(username))
			if b == nil {
				return fmt.Errorf("the user doesn't have any buckets")
			}
		} else if prefix != nil { // единственный некорректный случай
			return fmt.Errorf("empty username and non-empty prefix")
		}

		b = openBucketByPath(prefix, b)
		if b == nil {
			return fmt.Errorf("incorrect path")
		}

		value = b.Get(key)
		return nil
	})
	return value, err
}

func (b *Bolt) Set(username string, prefix []string, key, value []byte, bucketName []byte) error {
	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName) // kv
		if b == nil {
			return fmt.Errorf("failed to lookup bolt DB")
		}

		if username != "" {
			var err error
			b, err = b.CreateBucketIfNotExists([]byte(username))
			if err != nil {
				return err
			}
		} else if prefix != nil { // единственный некорректный случай
			return fmt.Errorf("empty username and non-empty prefix")
		}

		var err error
		fmt.Println(len(prefix))
		for _, pathPart := range prefix {
			fmt.Println("CURR:", pathPart)
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

func (b *Bolt) Delete(username string, prefix []string, key []byte, bucketName []byte) error {
	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName) // kv
		if b == nil {
			return fmt.Errorf("failed to lookup bolt DB")
		}

		if username != "" {
			b = b.Bucket([]byte(username))
		} else if prefix != nil { // единственный некорректный случай
			return fmt.Errorf("empty username and non-empty prefix")
		}

		if b == nil {
			return fmt.Errorf("the user doesn't have any buckets")
		}

		b = openBucketByPath(prefix, b)
		if b == nil {
			return fmt.Errorf("incorrect path")
		}

		return b.Delete(key)
	})
}

func (b *Bolt) ListKV(username []byte, path []string, vch chan Record, bch chan string, done chan interface{}) error {
	return b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(kvBucketName)
		if b == nil {
			return fmt.Errorf("failed to lookup bolt DB")
		}

		b = b.Bucket(username)
		if b == nil {
			return fmt.Errorf("failed to lookup user bucket")
		}

		b = openBucketByPath(path, b)

		if b == nil {
			return fmt.Errorf("bucket not found")
		}

		err := b.ForEach(func(k, v []byte) error {
			select {
			case <-done:
				return nil
			default:
				if v == nil {
					bch <- string(k)
				} else {
					vch <- Record{k, v}
				}
			}
			return nil
		})

		done <- 1

		return err
	})
}

func iterateBucketDecrypt(b *bolt.Bucket, indent string, ch chan RecordOut) error {
	if b == nil {
		return nil
	}

	return b.ForEach(func(k, v []byte) error {
		if v == nil {
			fmt.Printf("%sBucket: %s\n", indent, k)
			return iterateBucketDecrypt(b.Bucket(k), indent+"  ", ch)
		} else {
			ch <- RecordOut{Record{k, v}, indent}
		}
		return nil
	})
}

func (b *Bolt) ShowList(ch chan RecordOut, done chan interface{}) ([]Record, error) {
	fmt.Println("------KV------")
	var list []Record
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(kvBucketName)
		if b == nil {
			return fmt.Errorf("failed to lookup bolt DB")
		}
		err := iterateBucketDecrypt(b, "", ch)
		done <- 1
		close(ch)

		return err
	})

	if err != nil {
		return nil, err
	}
	fmt.Println("-----USERS-----")
	err = b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(userBucketName)
		if b == nil {
			return fmt.Errorf("failed to lookup bolt DB")
		}

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
		}
		return nil
	})
	return list, err
}
