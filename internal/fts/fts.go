package fts

import (
	"time"
)

type SearchResult struct {
	Total uint64
	Hits  []string
}

type Indexer interface {
	Open(path string) (*Indexer, error)
	Create(path, language string) (*Indexer, error)
	Close() error
	AddBooks(books []*Book, partial bool) error
	SearchByField(field, s string, page, pageSize int) (*SearchResult, error)
	GetMostRecentBooks(count int) ([]string, error)
}

type Book struct {
	LibId    string
	Title    string
	Authors  string
	Series   string
	SeriesNo int
	PubDate  time.Time
}

func (b *Book) BleveType() string {
	return "book"
}
