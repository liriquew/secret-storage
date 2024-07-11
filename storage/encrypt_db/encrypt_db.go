package encrypt_db

import (
	"crypto/rand"
	"errors"
	"fmt"
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
	rootName       = []byte("root")

	ErrUserNotFound     = errors.New("user not found")
	ErrWrongPassword    = errors.New("wrong password")
	ErrUserAlreadyExist = errors.New("user already exist")
	ErrKeyNotFound      = errors.New("key not found")
)

// Возращает хранилище, но без механизмов шифрования
func NewEncryptKV(path string) (*DBinfo, error) {
	db, err := bolt.New(path)

	if err != nil {
		return nil, err
	}

	rootPass, err := GeneratePassword(12)
	if err != nil {
		return nil, err
	}

	err = db.Set(nil, rootName, rootPass, userBucketName)
	if err != nil {
		return nil, err
	}

	return &DBinfo{
		Storage: &BoltEncrypt{
			db: db,
		},
		RootPass: string(rootPass),
	}, nil
}

// Устанавливает механизмы щифрования в хранилище
func (b *BoltEncrypt) InitWrapper(secterParts [][]byte) error {
	wrapper, err := encrypt.NewWrapper(b.db, secterParts)

	if err != nil {
		return err
	}

	b.wrapper = wrapper

	return nil
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

func makePrefix(username string, prefix []string) []string {
	prefix = append(prefix, "")
	copy(prefix[1:], prefix)
	prefix[0] = username
	return prefix
}

// Возвращает расшифрованное значение по ключу key
// в пространстве имен пользователя по пути path
func (b *BoltEncrypt) Get(username string, prefix []string, key string, bName []byte) ([]byte, error) {
	if b.wrapper == nil {
		return nil, fmt.Errorf("encrypt db not init full, need unseal")
	}
	if username != "" {
		prefix = makePrefix(username, prefix)
	}

	valueE, err := b.db.Get(prefix, []byte(key), bName)

	if err != nil {
		return nil, err
	}

	if valueE == nil {
		return nil, ErrKeyNotFound
	}

	return b.wrapper.Decrypt(valueE)
}

// Шифрует и устанавливает значение по ключу key
// в пространстве имен пользователя по пути path
func (b *BoltEncrypt) Set(username string, prefix []string, key, value string, bName []byte) error {
	if b.wrapper == nil {
		return fmt.Errorf("encrypt db not init full, need unseal")
	}
	valueE, err := b.wrapper.Encrypt([]byte(value))
	if err != nil {
		return err
	}
	if username != "" {
		prefix = makePrefix(username, prefix)
	} else if len(prefix) == 0 && string(bName) == string(kvBucketName) {
		return fmt.Errorf("not allowed top-level root keys (try .../root/...)")
	}

	return b.db.Set(prefix, []byte(key), []byte(valueE), bName)
}

// Удаляет ключ key, и связанное с ним значение
// в пространстве имен пользователя по пути path
func (b *BoltEncrypt) Delete(username string, prefix []string, key string, bName []byte) (int, error) {
	if b.wrapper == nil {
		return 0, fmt.Errorf("encrypt db not init full, need unseal")
	}
	if username != "" {
		prefix = makePrefix(username, prefix)
	}
	return b.db.Delete(prefix, []byte(key), bName)
}

// Возвращает список бакетов и ключей
// в пространстве имен пользователя по пути path
func (b *BoltEncrypt) List(username string, prefix []string) (*BucketInfo, error) {
	if b.wrapper == nil {
		return nil, fmt.Errorf("encrypt db not init full, need unseal")
	}
	if username != "" {
		prefix = makePrefix(username, prefix)
	}
	info, err := b.db.ListKV(prefix)
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

func (b *BoltEncrypt) prepareBuckets(bucket *bolt.BucketFullInfo) (*BucketFullInfo, error) {
	cur := &BucketFullInfo{}
	cur.Buckets = make(map[string]*BucketFullInfo, len(bucket.Buckets))
	cur.KVs = make([]KV, len(bucket.KVS))

	for i, kv := range bucket.KVS {
		valDec, err := b.wrapper.Decrypt(kv.Value)
		if err != nil {
			return nil, err
		}
		cur.KVs[i] = KV{string(kv.Key), string(valDec)}
	}

	var err error
	for bucketName, bucketInfo := range bucket.Buckets {
		cur.Buckets[bucketName], err = b.prepareBuckets(bucketInfo)
		if err != nil {
			return nil, err
		}
	}
	return cur, nil
}

// Возвращает "полное" представление бакета
// рекурсивно перечисляет все вложенные бакеты
func (b *BoltEncrypt) ListEncrypted(username string, prefix []string) (*BucketFullInfo, error) {
	if b.wrapper == nil {
		return nil, fmt.Errorf("encrypt db not init full, need unseal")
	}
	if username != "" {
		prefix = makePrefix(username, prefix)
	}
	BucketInfo, err := b.db.ShowBucketRecursion(prefix, kvBucketName)

	if err != nil {
		return nil, err
	}

	return b.prepareBuckets(BucketInfo)
}

func (b *BoltEncrypt) CreateNewUser(username, password string) error {
	isExist, err := b.CheckUser(username)

	if err != nil {
		return err
	}

	if isExist {
		return ErrUserAlreadyExist
	}

	return b.db.Set(nil, []byte(username), []byte(password), userBucketName)
}

func (b *BoltEncrypt) SelectUser(username, password string) error {
	dbPass, err := b.db.Get(nil, []byte(username), userBucketName)

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
	dbPass, err := b.db.Get(nil, []byte(username), userBucketName)

	return dbPass != nil, err
}
