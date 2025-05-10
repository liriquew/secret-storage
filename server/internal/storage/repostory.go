package storage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/liriquew/secret_storage/server/internal/models"
	"go.etcd.io/bbolt"
	bolt "go.etcd.io/bbolt"
)

var (
	recordsBucketName = []byte("kv")
	userBucketName    = []byte("user")
	metaBucketName    = []byte("meta")
	metaTokenName     = []byte("token")
)

var (
	ErrEmptyPathPart         = errors.New("empty path part")
	ErrBucketNotFound        = errors.New("bucket not found")
	ErrFailedToOpenTopBucket = errors.New("error while opening top-level bucket")
	ErrIncorrectPath         = errors.New("incorrect path")
	ErrIteratingBucket       = errors.New("error while iterating bucket")
)

func openBucketByPath(path []string, bucket *bolt.Bucket) (*bolt.Bucket, error) {
	for _, pathPart := range path {
		if pathPart == "" {
			return nil, ErrEmptyPathPart
		}
		bucket = bucket.Bucket([]byte(pathPart))
		if bucket == nil {
			return nil, fmt.Errorf("%w: bucket name - %s", ErrBucketNotFound, pathPart)
		}
	}
	return bucket, nil
}

func (s *Storage) Set(path []string, key string, value []byte, bucketName []byte) error {
	s.m.Lock()
	defer s.m.Unlock()

	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return ErrFailedToOpenTopBucket
		}

		var err error
		for _, pathPart := range path {
			b, err = b.CreateBucketIfNotExists([]byte(pathPart))
			if err != nil {
				return fmt.Errorf("error while creating path buckets (pathPart - %s): %w", pathPart, err)
			}
		}

		if b == nil {
			return fmt.Errorf("%w: path - %s", ErrIncorrectPath, strings.Join(path, "/"))
		}

		return b.Put([]byte(key), value)
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) Get(path []string, key string, bucketName []byte) ([]byte, error) {
	s.m.RLock()
	defer s.m.RUnlock()
	var value []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return ErrFailedToOpenTopBucket
		}

		b, err := openBucketByPath(path, b)
		if err != nil {
			return err
		}

		value = b.Get([]byte(key))
		return nil
	})
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (s *Storage) Delete(path []string, key string, bucketName []byte) (int, error) {
	deletedBuckets := 0
	s.m.Lock()
	defer s.m.Unlock()

	err := s.db.Update(func(tx *bolt.Tx) error {
		topLevelBucket := tx.Bucket([]byte(bucketName))
		if topLevelBucket == nil {
			return ErrFailedToOpenTopBucket
		}

		var dfs func(*bbolt.Bucket, int) (bool, error)
		dfs = func(b *bbolt.Bucket, pathIdx int) (bool, error) {
			// pathIdx is next path part to open
			if pathIdx == len(path) {
				err := b.Delete([]byte(key))
				if err != nil {
					return false, fmt.Errorf("error while deleting key: %w", err)
				}
				someKey, _ := b.Cursor().First()
				return someKey == nil, nil
			}

			nextB := b.Bucket([]byte(path[pathIdx]))
			if nextB == nil {
				return false, fmt.Errorf("%w: bucket name - %s", ErrBucketNotFound, path[pathIdx])
			}

			empty, err := dfs(nextB, pathIdx+1)
			if err != nil {
				return false, err
			}

			if !empty {
				return false, err
			}

			err = b.DeleteBucket([]byte(path[pathIdx]))
			if err != nil {
				return false, fmt.Errorf("error while deleting empty bucket: %w", err)
			}

			someKey, _ := b.Cursor().First()
			return someKey == nil, nil
		}

		if _, err := dfs(topLevelBucket, 0); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return deletedBuckets, nil
}

func (s *Storage) ListRecords(path []string) (*models.BucketInfo, error) {
	var bucketInfo models.BucketInfo

	s.m.RLock()
	defer s.m.RUnlock()

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(recordsBucketName)
		if b == nil {
			return ErrFailedToOpenTopBucket
		}

		b, err := openBucketByPath(path, b)
		if err != nil {
			return err
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if v == nil {
				bucketInfo.Buckets = append(bucketInfo.Buckets, string(k))
			} else {
				bucketInfo.Records = append(bucketInfo.Records, &models.Record{Key: k, Value: v})
			}
		}

		return err
	})

	if err != nil {
		return nil, err
	}

	return &bucketInfo, nil
}

func (s *Storage) ListRecordsRecursively(path []string) (*models.BucketFullInfo, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	BucketInfo := &models.BucketFullInfo{}
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(recordsBucketName)
		if b == nil {
			return ErrFailedToOpenTopBucket
		}

		b, err := openBucketByPath(path, b)
		if err != nil {
			return err
		}

		// bfs
		var iterateBucket func(*bbolt.Bucket) (*models.BucketFullInfo, error)
		iterateBucket = func(b *bbolt.Bucket) (*models.BucketFullInfo, error) {
			if b == nil {
				return nil, nil
			}

			bInfo := &models.BucketFullInfo{}

			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				if v == nil {
					subBucket := b.Bucket(k)
					subBucketInfo, err := iterateBucket(subBucket)
					if err != nil {
						return nil, fmt.Errorf("%w: bucket name - %s, err - %w", ErrIteratingBucket, string(k), err)
					}

					subBucketInfo.Name = string(k)
					bInfo.Buckets = append(bInfo.Buckets, subBucketInfo)
				} else {
					bInfo.Records = append(bInfo.Records, &models.Record{Key: k, Value: v})
				}
			}

			return bInfo, nil
		}
		BucketInfo, err = iterateBucket(b)

		return err
	})

	if err != nil {
		return nil, err
	}

	return BucketInfo, nil
}
