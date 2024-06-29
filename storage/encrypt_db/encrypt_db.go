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

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SecretInfo struct {
	Parts     int `json:"parts"`
	Threshold int `json:"threshold"`
}

type BucketInfo struct {
	Buckets []string `json:"buckets"`
	KVs     []KV     `json:"kvs"`
}

var (
	kvBucketName   = []byte("kv")
	userBucketName = []byte("user")
	rootBucketName = []byte("meta")
	rootTokenName  = []byte("token")
)

var (
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

func (b *BoltEncrypt) Set(username string, prefix []string, key, value string) error {
	valueE, err := b.wrapper.Encrypt([]byte(value))
	if err != nil {
		return err
	}
	return b.db.Set(username, prefix, []byte(key), []byte(valueE), kvBucketName)
}

func (b *BoltEncrypt) Delete(username string, prefix []string, key string) error {
	return b.db.Delete(username, prefix, []byte(key), kvBucketName)
}

func (b *BoltEncrypt) List(username string, prefix []string) (*BucketInfo, error) {
	valCh := make(chan bolt.Record)
	buckCh := make(chan string)
	done := make(chan interface{})

	var valList []KV
	var bucketList []string

	go func() {
		for {
			select {
			case v := <-valCh:
				decV, err := b.wrapper.Decrypt(v.Value)
				if err != nil {
					done <- 1
					return
				}
				valList = append(valList, KV{string(v.Key), string(decV)})
			case b := <-buckCh:
				bucketList = append(bucketList, b)
			case <-done:
				return
			}
		}
	}()

	err := b.db.ListKV([]byte(username), prefix, valCh, buckCh, done)
	if err != nil {
		return nil, err
	}

	return &BucketInfo{bucketList, valList}, nil
}

func (b *BoltEncrypt) ListEncrypted() {
	ch := make(chan bolt.RecordOut)
	done := make(chan interface{})

	go func() {
		for {
			select {
			case r := <-ch:
				valD, err := b.wrapper.Decrypt(r.KV.Value)
				if err != nil {
					fmt.Println("ERROR:\t\n", r.KV.Key, err)
					continue
				}
				fmt.Printf("%sINFO:    key=%s, value=%s\n", r.Indent, r.KV.Key, valD)
			case <-done:
				return
			}
		}
	}()

	_, err := b.db.ShowList(ch, done)

	if err != nil {
		fmt.Println(err)
		return
	}
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
