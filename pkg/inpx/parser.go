package inpx

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Defines the type of particular field in the record.
type field int

// Known types of field.
const (
	unknown field = iota
	author
	genre
	title
	series
	seriesNo
	fileName
	fileSize
	libId
	deleted
	ext
	publishedDate
	language
	libRate
	keywords
	insNo
	folder
)

var (
	fieldsMap = map[string]field{
		"AUTHOR":   author,
		"GENRE":    genre,
		"TITLE":    title,
		"SERIES":   series,
		"SERNO":    seriesNo,
		"FILE":     fileName,
		"SIZE":     fileSize,
		"LIBID":    libId,
		"DEL":      deleted,
		"EXT":      ext,
		"DATE":     publishedDate,
		"INSNO":    insNo,
		"FOLDER":   folder,
		"LANG":     language,
		"LIBRATE":  libRate,
		"KEYWORDS": keywords,
	}
)

// defaultStructure is a fallback for files without `structure.info`.
var defaultStructure = []field{
	author, genre, title, series, seriesNo, fileName, fileSize, libId, deleted, ext, publishedDate, language, libRate, keywords,
}

var (
	extraTrimChars = "\uFEFF"
)

// Author represents the names of author.
type Author struct {
	LastName   string
	FirstName  string
	MiddleName string
}

// File represents book file information.
type File struct {
	Name    string
	Size    int
	Ext     string
	Folder  string
	Archive string
}

// Book represents a book record in library.
type Book struct {
	Authors       []Author
	Genres        []string
	Title         string
	Series        string
	SeriesNo      int
	File          File
	LibId         int
	Deleted       bool
	PublishedDate time.Time
	Language      string
}

// Parser represents `.inpx` collection.
type Parser struct {
	arc       *zip.ReadCloser
	bookCh    chan *Book
	structure []field
	err       error
	Name      string
	Id        int
	Comment   string
	Version   string
}

// Open reads collection meta info and validates it.
func Open(name string) (*Parser, error) {
	arc, err := zip.OpenReader(name)
	if err != nil {
		return nil, fmt.Errorf("error opening archive: %v", err)
	}

	r := new(Parser)
	r.arc = arc
	r.bookCh = make(chan *Book, 128)

	err = r.readCollection()
	if err != nil {
		_ = r.arc.Close()
		return nil, fmt.Errorf("error parsing archive: %v", err)
	}

	err = r.readVersion()
	if err != nil {
		_ = r.arc.Close()
		return nil, fmt.Errorf("error parsing archive: %v", err)
	}

	r.structure = r.getStructure()

	return r, nil
}

// Close underlying zip archive.
func (p *Parser) Close() {
	_ = p.arc.Close()
}

// Err returns the most recent decoder error if any, or nil.
func (p *Parser) Err() error { return p.err }

// Stream begins parsing from the underlying reader and returns a streaming Book channel.
func (p *Parser) Stream() chan *Book {
	go p.parse()
	return p.bookCh
}

// parse decodes records from files in archives and emits it to Book channel.
func (p *Parser) parse() {
	defer close(p.bookCh)

	for _, f := range p.arc.File {
		if !strings.HasSuffix(f.Name, ".inp") {
			continue
		}

		archive := strings.TrimSuffix(f.Name, ".inp")
		rc, err := f.Open()
		if err != nil {
			p.err = err
			break
		}

		br := bufio.NewReader(rc)
		for {
			line, err := br.ReadBytes('\n')
			if len(line) == 0 && err == io.EOF {
				break
			}

			if err != nil && err != io.EOF {
				rc.Close()
				p.err = err
				return
			}

			line = bytes.TrimSpace(line)
			values := bytes.Split(line, []byte{0x04})
			book, err := mapFieldsToBook(p.structure, values[:len(values)-1])
			if err != nil {
				rc.Close()
				p.err = err
				return
			}

			book.File.Archive = archive
			p.bookCh <- book
		}
		rc.Close()
	}
}

// Splits string separated by `:` and cleans from empty trailing element.
func splitColonStr(str string) []string {
	values := strings.Split(str, ":")
	values = values[:len(values)-1]
	return values
}

func getAuthorFromStr(s string) Author {
	names := strings.Split(s, ",")
	switch len(names) {
	case 1:
		return Author{
			LastName: names[0],
		}
	case 2:
		return Author{
			LastName:  names[0],
			FirstName: names[1],
		}
	default:
		return Author{
			LastName:   names[0],
			FirstName:  names[1],
			MiddleName: names[2],
		}
	}
}

// Constructs Book from array of fields.
func mapFieldsToBook(structure []field, values [][]byte) (*Book, error) {
	if len(structure) != len(values) {
		return nil, fmt.Errorf("fields count doesn't math with a structure. expected %d, got %d", len(structure), len(values))
	}

	book := new(Book)
	for i, f := range structure {
		value := string(values[i])
		switch f {
		case author:
			authors := splitColonStr(value)
			if len(authors) == 1 && authors[0] == "" {
				continue
			}
			for _, author := range authors {
				book.Authors = append(book.Authors, getAuthorFromStr(author))
			}
		case genre:
			book.Genres = splitColonStr(value)
		case title:
			book.Title = value
		case series:
			book.Series = value
		case seriesNo:
			v, err := strconv.Atoi(value)
			if err == nil {
				book.SeriesNo = v
			}
		case libId:
			v, err := strconv.Atoi(value)
			if err == nil {
				book.LibId = v
			}
		case deleted:
			v, err := strconv.ParseBool(value)
			if err == nil {
				book.Deleted = v
			}
		case language:
			book.Language = value
		case publishedDate:
			v, err := time.Parse("2006-01-02", value)
			if err == nil {
				book.PublishedDate = v
			}
		case fileSize:
			v, err := strconv.Atoi(value)
			if err == nil {
				book.File.Size = v
			}
		case fileName:
			book.File.Name = value
		case ext:
			book.File.Ext = value
		case folder:
			book.File.Folder = value
		}
	}
	return book, nil
}

// Reads collection version info.
func (p *Parser) readVersion() error {
	r, err := p.getFileByName("version.info")
	if err != nil {
		return fmt.Errorf("version.info not found: %v", err)
	}
	defer r.Close()

	br := bufio.NewReader(r)
	p.Version, err = readCleanString(br)
	if err != nil {
		return err
	}

	return nil
}

// Reads collection meta.
func (p *Parser) readCollection() error {
	r, err := p.getFileByName("collection.info")
	if err != nil {
		return fmt.Errorf("collection.info not found: %v", err)
	}
	defer r.Close()

	br := bufio.NewReader(r)
	p.Name, err = readCleanString(br)
	if err != nil {
		return err
	}

	// filename
	_, err = br.ReadString('\n')

	_, err = fmt.Fscan(br, &p.Id)
	if err != nil {
		return err
	}

	// FIXME: there is can be more than one line, but who cares?
	_, _ = br.ReadString('\n')
	p.Comment, err = readCleanString(br)

	return nil
}

// Parses structure from archive and returns it.
func (p *Parser) getStructure() []field {
	r, err := p.getFileByName("structure.info")
	if err != nil {
		return defaultStructure
	}
	defer r.Close()

	br := bufio.NewReader(r)
	ssv, err := readCleanString(br)
	if err != nil {
		return defaultStructure
	}

	structure := make([]field, 0)
	fieldNames := strings.Split(ssv, ";")
	for _, fieldName := range fieldNames[:len(fieldNames)-1] {
		structure = append(structure, parseFieldName(fieldName))
	}

	return structure
}

// Opens file from archive by file name.
func (p *Parser) getFileByName(name string) (io.ReadCloser, error) {
	for _, f := range p.arc.File {
		if f.Name != name {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, err
		}

		return rc, err
	}

	return nil, fmt.Errorf("file not found: %s", name)
}

// Read line from reader and returns it trimmed.
func readCleanString(r *bufio.Reader) (string, error) {
	s, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	
	return strings.Trim(strings.TrimSpace(s), extraTrimChars), nil
}

// Maps string representation of field name to actual value.
func parseFieldName(field string) field {
	c, ok := fieldsMap[field]
	if ok {
		return c
	} else {
		return unknown
	}
}
