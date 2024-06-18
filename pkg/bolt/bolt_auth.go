package bolt

import (
	"errors"
	bolt "go.etcd.io/bbolt"
)

func (b *Bolt) CreateNewUser(username, password string) error {
	var bPassword []byte
	b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(userBucketName)
		bPassword = b.Get([]byte(username))
		return nil
	})
	if bPassword != nil {
		return errors.New("user already exists")
	}
	return b.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(userBucketName)
		return b.Put([]byte(username), []byte(password))
	})
}

func (b *Bolt) SelectUser(username, password string) error {
	var bPassword []byte
	b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(userBucketName)
		bPassword = b.Get([]byte(username))
		return nil
	})
	if bPassword == nil {
		return errors.New("user not found")
	}

	if password != string(bPassword) {
		return errors.New("wrong password")
	}

	return nil
}

func (b *Bolt) CheckUser(username string) error {
	var bPassword []byte
	b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(userBucketName)
		bPassword = b.Get([]byte(username))
		return nil
	})
	if bPassword == nil {
		return errors.New("user not found")
	}
	return nil
}
