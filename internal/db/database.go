package db

import (
	"path/filepath"
	"strings"

	"github.com/shemanaev/inpxer/internal/db/badgerstore"
	"github.com/shemanaev/inpxer/internal/db/storer"
	"github.com/shemanaev/inpxer/internal/fts"
	"github.com/shemanaev/inpxer/internal/fts/blevefts"
	"github.com/shemanaev/inpxer/internal/model"
)

const (
	blevePath = "bleve"
	boltPath  = "badger"
)

type SearchResult struct {
	Total uint64
	Hits  []*model.Book
}

type Store struct {
	fts fts.Indexer
	db  storer.BookStorer
}

func Open(path string) (*Store, error) {
	indexer, err := blevefts.Open(filepath.Join(path, blevePath))
	if err != nil {
		return nil, err
	}

	data, err := badgerstore.Open(filepath.Join(path, boltPath))
	if err != nil {
		return nil, err
	}

	return &Store{
		fts: indexer,
		db:  data,
	}, nil
}

func Create(path, language string) (*Store, error) {
	indexer, err := blevefts.Create(filepath.Join(path, blevePath), language)
	if err != nil {
		return nil, err
	}

	data, err := badgerstore.Open(filepath.Join(path, boltPath))
	if err != nil {
		return nil, err
	}

	return &Store{
		fts: indexer,
		db:  data,
	}, nil
}

func (s *Store) Close() error {
	err := s.fts.Close()
	if err != nil {
		s.db.Close()
		return err
	}
	return s.db.Close()
}

func (s *Store) AddBooks(books []*model.Book, partial bool) error {
	err := s.db.AddBooks(books, partial)
	if err != nil {
		return err
	}

	var ftsBooks []*fts.Book
	for _, book := range books {
		ftsBooks = append(ftsBooks, ftsBookFromModel(book))
	}
	err = s.fts.AddBooks(ftsBooks, partial)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) GetBookById(id string) (*model.Book, error) {
	return s.db.GetBookById(id)
}

func (s *Store) SearchByField(field, query string, page, pageSize int) (*SearchResult, error) {
	search, err := s.fts.SearchByField(field, query, page, pageSize)
	if err != nil {
		return nil, err
	}

	var books []*model.Book
	for _, id := range search.Hits {
		book, err := s.db.GetBookById(id)
		if err == nil {
			books = append(books, book)
		}
	}

	return &SearchResult{
		Total: search.Total,
		Hits:  books,
	}, nil
}

func (s *Store) GetMostRecentBooks(count int) ([]*model.Book, error) {
	search, err := s.fts.GetMostRecentBooks(count)
	if err != nil {
		return nil, err
	}

	var books []*model.Book
	for _, id := range search {
		book, err := s.db.GetBookById(id)
		if err == nil {
			books = append(books, book)
		}
	}

	return books, nil
}

func ftsBookFromModel(book *model.Book) *fts.Book {
	authors := make([]string, len(book.Authors))
	for i, v := range book.Authors {
		authors[i] = v.String()
	}

	return &fts.Book{
		LibId:    book.LibId,
		Title:    book.Title,
		Authors:  strings.Join(authors, ","),
		Series:   book.Series,
		SeriesNo: book.SeriesNo,
		PubDate:  book.PubDate,
	}
}
