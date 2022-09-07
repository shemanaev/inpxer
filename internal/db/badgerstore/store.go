package badgerstore

import (
	"bytes"
	"encoding/gob"

	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"

	"github.com/shemanaev/inpxer/internal/db/storer"
	"github.com/shemanaev/inpxer/internal/model"
)

type Database struct {
	storer.BookStorer
	db *badger.DB
}

func Open(path string) (*Database, error) {
	opts := badger.DefaultOptions(path).
		WithLoggingLevel(badger.WARNING).
		WithCompression(options.None).
		WithBlockCacheSize(0).
		WithNumMemtables(1).
		WithValueLogFileSize(64 << 20) // 64 Mb
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) AddBooks(books []*model.Book, partial bool) error {
	err := d.db.Update(func(tx *badger.Txn) error {
		for _, v := range books {
			if partial {
				if _, err := tx.Get([]byte(v.LibId)); err == nil {
					continue
				}
			}

			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			err := enc.Encode(v)
			if err != nil {
				return err
			}

			err = tx.Set([]byte(v.LibId), buf.Bytes())
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
	err := d.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get([]byte(id))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			buf := bytes.NewBuffer(val)
			dev := gob.NewDecoder(buf)

			err := dev.Decode(&result)
			if err != nil {
				return err
			}

			return nil
		})

		return err
	})

	return &result, err
}
