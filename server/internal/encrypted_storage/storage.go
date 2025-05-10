package encryptedstorage

import (
	"github.com/liriquew/secret_storage/server/internal/lib/config"
	"github.com/liriquew/secret_storage/server/internal/lib/encrypt"
	"github.com/liriquew/secret_storage/server/internal/models"
	"github.com/liriquew/secret_storage/server/internal/storage"
)

type Storage interface {
	Get(path []string, key string, bucketName []byte) ([]byte, error)
	Set(path []string, key string, value []byte, bucketName []byte) error
	Delete(path []string, key string, bucketName []byte) (int, error)
	ListRecords(path []string) (*models.BucketInfo, error)
	ListRecordsRecursively(path []string) (*models.BucketFullInfo, error)
}

type Erypter interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

type EncryptedStorage struct {
	db      Storage
	crypter Erypter
}

func New(cfg config.StorageConfig, key []byte) (*EncryptedStorage, error) {
	db, err := storage.New(cfg)
	if err != nil {
		return nil, err
	}

	crypter, err := encrypt.NewEncrypter(key)
	if err != nil {
		return nil, err
	}

	return &EncryptedStorage{
		db:      db,
		crypter: crypter,
	}, nil
}
