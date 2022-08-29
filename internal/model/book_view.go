package model

import (
	"strconv"
	"strings"
	"time"
)

type BookView struct {
	LibId    string
	Title    string
	Authors  []string
	Genres   []string
	Series   string
	SeriesNo int
	File     string
	Ext      string
	Size     int
	PubDate  time.Time
	PubYear  string
	Language string
}

func NewBookView(src *Book) *BookView {
	pubYear := strconv.Itoa(src.PubDate.Year())
	genres := strings.Split(src.GenresStored, ":")
	authors := strings.Split(src.Authors, ",")
	return &BookView{
		LibId:    src.LibId,
		Title:    src.Title,
		Authors:  authors,
		Genres:   genres,
		Series:   src.Series,
		SeriesNo: src.SeriesNo,
		File:     src.File,
		Ext:      src.Ext,
		Size:     src.Size,
		PubDate:  src.PubDate,
		PubYear:  pubYear,
		Language: src.Language,
	}
}
