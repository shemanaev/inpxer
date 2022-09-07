package storer

import "github.com/shemanaev/inpxer/internal/model"

type BookStorer interface {
	Open(path string) (*BookStorer, error)
	Close() error
	AddBooks(books []*model.Book, partial bool) error
	GetBookById(id string) (*model.Book, error)
}
