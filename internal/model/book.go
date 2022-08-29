package model

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shemanaev/inpxer/pkg/inpx"
)

type Book struct {
	LibId        string
	Title        string
	Authors      string
	GenresStored string
	Series       string
	SeriesNo     int
	Folder       string
	File         string
	Ext          string
	Archive      string
	Size         int
	PubDate      time.Time
	Language     string
}

func (b *Book) BleveType() string {
	return "book"
}

var seriesSuffixes = []string{"[a]", "[p]", "[m]"}

func NewBook(book *inpx.Book) *Book {
	authors := ""
	for _, a := range book.Authors {
		var name string
		if a.FirstName == "" && a.MiddleName == "" {
			name = a.LastName
		} else if a.MiddleName == "" {
			name = fmt.Sprintf("%s %s", a.FirstName, a.LastName)
		} else {
			name = fmt.Sprintf("%s %s %s", a.FirstName, a.MiddleName, a.LastName)
		}
		authors = authors + name + ","
	}
	if len(authors) > 0 {
		authors = authors[:len(authors)-1]
	}

	genres := ""
	for _, g := range book.Genres {
		genres = genres + g + ":"
	}
	if len(genres) > 0 {
		genres = genres[:len(genres)-1]
	}

	series := book.Series
	for _, suffix := range seriesSuffixes {
		series = strings.TrimSuffix(series, suffix)
	}

	return &Book{
		LibId:        strconv.Itoa(book.LibId),
		Title:        book.Title,
		Authors:      authors,
		Series:       series,
		GenresStored: genres,
		SeriesNo:     book.SeriesNo,
		Folder:       filepath.ToSlash(book.File.Folder),
		File:         book.File.Name,
		Ext:          book.File.Ext,
		Archive:      book.File.Archive,
		Size:         book.File.Size,
		PubDate:      book.PublishedDate,
		Language:     book.Language,
	}
}
