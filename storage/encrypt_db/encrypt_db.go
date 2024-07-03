package encrypt_db

import (
	"crypto/rand"
	"errors"
	"secret-storage/storage/bolt"
	"secret-storage/storage/encrypt"
	"secret-storage/storage/shamir"
)

type BoltEncrypt struct {
	db      *bolt.Bolt
	wrapper *encrypt.Wrapper
}

type SecretInfo struct {
	Parts     int `json:"parts"`
	Threshold int `json:"threshold"`
}

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type BucketInfo struct {
	Buckets []string `json:"buckets"`
	KVs     []KV     `json:"kvs"`
}

type BucketFullInfo struct {
	Buckets map[string]*BucketFullInfo `json:"buckets"`
	KVs     []KV                       `json:"kvs"`
}

var (
	kvBucketName   = []byte("kv")
	userBucketName = []byte("user")
	rootBucketName = []byte("meta")
	rootTokenName  = []byte("token")

	ErrUserNotFound     = errors.New("user not found")
	ErrWrongPassword    = errors.New("wrong password")
	ErrUserAlreadyExist = errors.New("user already exist")
)

func NewEncryptKV(path string, secterParts [][]byte) (*BoltEncrypt, error) {
	wrapper, err := encrypt.NewWrapper(path, secterParts)

	if err != nil {
		return nil, err
	}

	db, err := bolt.New(path)

	if err != nil {
		return nil, err
	}
	return &BoltEncrypt{
		db:      db,
		wrapper: wrapper,
	}, nil
}

func (b *BoltEncrypt) Close() error {
	return b.db.Close()
}

func MakeMasterKey(keyInfo SecretInfo) ([][]byte, error) {
	masterKey := make([]byte, 32)
	_, err := rand.Read(masterKey)
	if err != nil {
		return nil, err
	}

	parts, err := shamir.Split(masterKey, keyInfo.Parts, keyInfo.Threshold)

	return parts, err
}

// Возвращает расшифрованное значение по ключу key
// в пространстве имен пользователя по пути path
func (b *BoltEncrypt) Get(username string, prefix []string, key string) ([]byte, error) {
	valueE, err := b.db.Get(username, prefix, []byte(key), kvBucketName)

	if err != nil {
		return nil, err
	}

	if valueE == nil {
		return nil, nil
	}

	valueD, err := b.wrapper.Decrypt(valueE)
	if err != nil {
		return nil, err
	}
	return valueD, nil
}

// Шифрует и устанавливает значение по ключу key
// в пространстве имен пользователя по пути path
func (b *BoltEncrypt) Set(username string, prefix []string, key, value string) error {
	valueE, err := b.wrapper.Encrypt([]byte(value))
	if err != nil {
		return err
	}
	return b.db.Set(username, prefix, []byte(key), []byte(valueE), kvBucketName)
}

// Удаляет ключ key, и связанное с ним значение
// в пространстве имен пользователя по пути path
func (b *BoltEncrypt) Delete(username string, prefix []string, key string) error {
	return b.db.Delete(username, prefix, []byte(key), kvBucketName)
}

// Возвращает список бакетов и ключей
// в пространстве имен пользователя по пути path
func (b *BoltEncrypt) List(username string, prefix []string) (*BucketInfo, error) {
	info, err := b.db.ListKV([]byte(username), prefix)
	if err != nil {
		return nil, err
	}

	infoDec := &BucketInfo{}
	infoDec.KVs = make([]KV, len(info.KVs))
	infoDec.Buckets = info.Buckets

	for i, kvInfo := range info.KVs {
		infoDec.KVs[i].Key = string(info.KVs[i].Key)
		kvInfo.Value, err = b.wrapper.Decrypt(kvInfo.Value)
		infoDec.KVs[i].Value = string(kvInfo.Value)
		if err != nil {
			return nil, err
		}
	}

	return infoDec, nil
}

func (b *BoltEncrypt) showBuckets(bucketName string, bucket *bolt.BucketFullInfo, indent string) (*BucketFullInfo, error) {
	// fmt.Printf("%sBUCKET: %s\n", indent, bucketName)
	cur := &BucketFullInfo{}
	cur.Buckets = make(map[string]*BucketFullInfo, len(bucket.Buckets))
	cur.KVs = make([]KV, len(bucket.KVS))

	for i, kv := range bucket.KVS {
		valDec, err := b.wrapper.Decrypt(kv.Value)
		if err != nil {
			return nil, err
		}
		cur.KVs[i] = KV{string(kv.Key), string(valDec)}
		// fmt.Printf("%sKEYVAL: %s : %s\n", indent, kv.Key, valDec)
	}

	var err error
	for bucketName, bucketInfo := range bucket.Buckets {
		cur.Buckets[bucketName], err = b.showBuckets(bucketName, bucketInfo, indent+"  ")
		if err != nil {
			return nil, err
		}
	}
	return cur, nil
}

func (b *BoltEncrypt) ListEncrypted(username string, prefix []string) (*BucketFullInfo, error) {
	BucketInfo, err := b.db.ShowBucketRecursion([]byte(username), prefix, kvBucketName)

	if err != nil {
		return nil, err
	}

	return b.showBuckets(string(kvBucketName), BucketInfo, "")
}

func (b *BoltEncrypt) GetRootToken() ([]byte, error) {
	return b.db.Get("", nil, rootTokenName, rootBucketName)
}

func (b *BoltEncrypt) CreateNewUser(username, password string) error {
	isExist, err := b.CheckUser(username)

	if err != nil {
		return err
	}

	if isExist {
		return ErrUserAlreadyExist
	}

	return b.db.Set("", nil, []byte(username), []byte(password), userBucketName)
}

func (b *BoltEncrypt) SelectUser(username, password string) error {
	dbPass, err := b.db.Get("", nil, []byte(username), userBucketName)

	if err != nil {
		return err
	}

	if dbPass == nil {
		return ErrUserNotFound
	}

	if password != string(dbPass) {
		return ErrWrongPassword
	}

	return nil
}

func (b *BoltEncrypt) CheckUser(username string) (bool, error) {
	dbPass, err := b.db.Get("", nil, []byte(username), userBucketName)

	return dbPass != nil, err
}
