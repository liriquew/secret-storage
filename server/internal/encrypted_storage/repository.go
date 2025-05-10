package encryptedstorage

import (
	"errors"

	"github.com/liriquew/secret_storage/server/internal/models"
	"github.com/liriquew/secret_storage/server/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

var (
	recordsBucketName = []byte("kv")
	userBucketName    = []byte("user")
	metaBucketName    = []byte("meta")
	metaTokenName     = []byte("token")
)

var (
	ErrEmptyPathPart   = errors.New("empty path part")
	ErrBucketNotFound  = errors.New("bucket not found")
	ErrIncorrectPath   = errors.New("incorrect path")
	ErrIteratingBucket = errors.New("error while iterating bucket")

	ErrRecordNotFound = errors.New("record not found")
)

func (es *EncryptedStorage) Set(path []string, key string, value []byte) error {
	value, err := es.crypter.Encrypt(value)
	if err != nil {
		return err
	}

	if err := es.db.Set(path, key, value, recordsBucketName); err != nil {
		if errors.Is(err, storage.ErrIncorrectPath) {
			return ErrIncorrectPath
		}
		return err
	}

	return nil
}

func (es *EncryptedStorage) Get(path []string, key string) ([]byte, error) {
	value, err := es.db.Get(path, key, recordsBucketName)
	if err != nil {
		if errors.Is(err, storage.ErrBucketNotFound) {
			return nil, ErrBucketNotFound
		}

		return nil, err
	}

	decryptedValue, err := es.crypter.Decrypt(value)
	if err != nil {
		return nil, err
	}

	if decryptedValue == nil {
		return nil, ErrRecordNotFound
	}

	return decryptedValue, err
}

func (es *EncryptedStorage) Delete(path []string, key string) (int, error) {
	deletedBuckets, err := es.db.Delete(path, key, recordsBucketName)
	if err != nil {
		if errors.Is(err, storage.ErrBucketNotFound) {
			return 0, ErrBucketNotFound
		}
		return 0, err
	}

	return deletedBuckets, nil
}

func (es *EncryptedStorage) ListRecords(path []string) (*models.BucketInfo, error) {
	bucketInfo, err := es.db.ListRecords(path)
	if err != nil {
		return nil, err
	}

	for i, record := range bucketInfo.Records {
		bucketInfo.Records[i].Value, err = es.crypter.Decrypt(record.Value)
		if err != nil {
			return nil, err
		}
	}

	return bucketInfo, nil
}

func (es *EncryptedStorage) ListRecordsRecursively(path []string) (*models.BucketFullInfo, error) {
	bucketFullInfo, err := es.db.ListRecordsRecursively(path)
	if err != nil {
		return nil, err
	}

	err = es.decryptBucketFullInfo(bucketFullInfo)
	if err != nil {
		return nil, err
	}

	return bucketFullInfo, err
}

func (es *EncryptedStorage) decryptBucketFullInfo(bucketInfo *models.BucketFullInfo) error {
	var err error
	for i, record := range bucketInfo.Records {
		bucketInfo.Records[i].Value, err = es.crypter.Decrypt(record.Value)
		if err != nil {
			return err
		}
	}

	for _, bucket := range bucketInfo.Buckets {
		if err := es.decryptBucketFullInfo(bucket); err != nil {
			return err
		}
	}

	return nil
}

func (es *EncryptedStorage) CreateUser(user *models.User) error {
	passHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
	if err != nil {
		return err
	}

	if err := es.db.Set(nil, user.Username, passHash, userBucketName); err != nil {
		return err
	}

	return nil
}

func (es *EncryptedStorage) CheckUserCredentials(user *models.User) error {
	passHash, err := es.db.Get(nil, user.Username, userBucketName)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword(passHash, []byte(user.Password)); err != nil {
		return err
	}

	return nil
}
