package boltstore

import (
	"bytes"
	"encoding/gob"

	bolt "go.etcd.io/bbolt"

	"github.com/shemanaev/inpxer/internal/model"
)

var BucketName = []byte("Books")

type Database struct {
	db *bolt.DB
}

func Open(path string) (*Database, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) AddBooks(books []*model.Book) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(BucketName)
		if err != nil {
			return err
		}

		for _, v := range books {
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			err := enc.Encode(v)
			if err != nil {
				return err
			}

			err = b.Put([]byte(v.LibId), buf.Bytes())
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func (d *Database) GetBookById(id string) (*model.Book, error) {
	var result model.Book
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BucketName)
		v := b.Get([]byte(id))

		buf := bytes.NewBuffer(v)
		dev := gob.NewDecoder(buf)

		err := dev.Decode(&result)
		if err != nil {
			return err
		}

		return nil
	})

	return &result, err
}
