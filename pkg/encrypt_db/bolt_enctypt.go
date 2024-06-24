package encrypt_db

import (
	"errors"
	"fmt"
	"kv-storage/pkg/bolt"
	"kv-storage/pkg/encrypt"
)

type BoltEncrypt struct {
	db      *bolt.Bolt
	wrapper *encrypt.Wrapper
}

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var (
	kvBucketName   = []byte("kv")
	userBucketName = []byte("user")
	rootBucketName = []byte("meta")
	rootTokenName  = []byte("token")
)

var (
	UserNotFoundErr     = errors.New("user not found")
	WrongPasswordErr    = errors.New("wrong password")
	UserAlreadyExistErr = errors.New("user already exist")
)

func NewEncryptKV(path string) (*BoltEncrypt, error) {
	wrapper, err := encrypt.NewWrapper(path)

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

func (b *BoltEncrypt) Get(key []byte) ([]byte, error) {
	valueE, err := b.db.Get(key, kvBucketName)

	if err != nil {
		return nil, err
	}

	if valueE == nil {
		return nil, nil
	}

	fmt.Println("Start Decode: ")
	valueD, err := b.wrapper.Decrypt(valueE)
	if err != nil {
		return nil, err
	}
	return valueD, nil
}

func (b *BoltEncrypt) Set(key, value []byte) error {
	valueE, err := b.wrapper.Encrypt(value)
	if err != nil {
		return err
	}
	return b.db.Set(key, valueE, kvBucketName)
}

func (b *BoltEncrypt) Delete(key []byte) error {
	return b.db.Delete(key, kvBucketName)
}

func (b *BoltEncrypt) List() ([]KV, error) {
	list, err := b.db.ListKV()
	if err != nil {
		return nil, err
	}

	listEnc := make([]KV, len(list))

	for i, v := range list {
		value, err := b.wrapper.Decrypt(v.Value)
		if err != nil {
			return nil, err
		}
		listEnc[i] = KV{string(v.Key), string(value)}
	}
	return listEnc, nil
}

func (b *BoltEncrypt) ListEncrypted() {
	list, err := b.db.ShowList()

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, r := range list {
		valD, err := b.wrapper.Decrypt(r.Value)
		if err != nil {
			fmt.Println("ERROR:\t\n", r.Key, err)
			continue
		}
		fmt.Printf("INFO:\tkey=%s, value=%s\n", r.Key, valD)
	}
}

func (b *BoltEncrypt) GetRootToken() ([]byte, error) {
	return b.db.Get(rootTokenName, rootBucketName)
}

func (b *BoltEncrypt) CreateNewUser(username, password string) error {
	isExist, err := b.CheckUser(username)

	if err != nil {
		return err
	}

	if isExist {
		return UserAlreadyExistErr
	}

	return b.db.Set([]byte(username), []byte(password), userBucketName)
}

func (b *BoltEncrypt) SelectUser(username, password string) error {
	dbPass, err := b.db.Get([]byte(username), userBucketName)

	if err != nil {
		return err
	}

	if dbPass == nil {
		return UserNotFoundErr
	}

	if password != string(dbPass) {
		return WrongPasswordErr
	}

	return nil

}

func (b *BoltEncrypt) CheckUser(username string) (bool, error) {
	dbPass, err := b.db.Get([]byte(username), userBucketName)

	return dbPass != nil, err
}
