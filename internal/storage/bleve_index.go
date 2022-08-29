package storage

import (
	"fmt"
	"log"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/lang/ru"
	"github.com/blevesearch/bleve/v2/index/scorch"
	"github.com/blevesearch/bleve/v2/index/upsidedown/store/boltdb"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/mitchellh/mapstructure"

	"github.com/shemanaev/inpxer/internal/model"
)

var (
	extractFieldsAll = []string{
		"LibId",
		"Title",
		"Authors",
		"GenresStored",
		"Series",
		"SeriesNo",
		"Folder",
		"File",
		"Ext",
		"Archive",
		"Size",
		"PubDate",
		"Language",
	}

	extractFieldsDownload = []string{
		"LibId",
		"Title",
		"Authors",
		"Folder",
		"File",
		"Ext",
		"Archive",
	}
)

type SearchResult struct {
	Total uint64
	Hits  []*model.BookView
}

type BleveIndex struct {
	index bleve.Index
}

func Open(path, language string, create bool) (*BleveIndex, error) {
	idx, err := bleve.Open(path)
	if err != nil {
		if !create {
			return nil, err
		}

		analyzer := getAnalyzer(language)
		idx, err = bleve.NewUsing(path, createBookMapping(analyzer), scorch.Name, boltdb.Name, nil)
		if err != nil {
			return nil, err
		}
	}

	res := BleveIndex{
		index: idx,
	}
	return &res, nil
}

func (i *BleveIndex) Close() {
	i.index.Close()
}

func (i *BleveIndex) Add(books []*model.Book) error {
	batch := i.index.NewBatch()
	for _, book := range books {
		err := batch.Index(book.LibId, book)
		if err != nil {
			log.Printf("Error index book %v, %v", book, err)
			return err
		}
	}

	err := i.index.Batch(batch)
	if err != nil {
		log.Printf("Error indexing batch %v", err)
		return err
	}

	return nil
}

func (i *BleveIndex) SearchByField(field string, s string, page int, pageSize int) (*SearchResult, error) {
	query := bleve.NewMatchQuery(s)
	query.SetField(field)
	search := bleve.NewSearchRequestOptions(query, pageSize, page*pageSize, false)
	search.Fields = extractFieldsAll

	switch field {
	case "Title":
		search.SortBy([]string{"-_score", "-PubDate", "-SeriesNo"})
	case "Authors":
		search.SortBy([]string{"-_score", "-PubDate"})
	case "Series":
		search.SortBy([]string{"-_score", "Series", "-SeriesNo"})
	}

	searchResults, err := i.index.Search(search)
	if err != nil {
		return nil, err
	}

	books := make([]*model.BookView, 0)
	for _, v := range searchResults.Hits {
		var book model.Book
		err := DecodeStruct(v.Fields, &book)
		if err != nil {
			return nil, err
		}

		books = append(books, model.NewBookView(&book))
	}

	res := SearchResult{
		Total: searchResults.Total,
		Hits:  books,
	}
	return &res, nil
}

func (i *BleveIndex) FindById(id string) (*model.Book, error) {
	query := bleve.NewTermQuery(id)
	query.SetField("_id")
	search := bleve.NewSearchRequest(query)
	search.Fields = extractFieldsDownload
	searchResults, err := i.index.Search(search)
	if err != nil {
		return nil, err
	}

	if searchResults.Total == 0 {
		return nil, fmt.Errorf("book with id %s not found", id)
	}

	var book model.Book
	err = mapstructure.Decode(searchResults.Hits[0].Fields, &book)
	if err != nil {
		return nil, err
	}

	return &book, nil
}

func createBookMapping(analyzer string) *mapping.IndexMappingImpl {
	bookMapping := bleve.NewDocumentMapping()

	indexedText := bleve.NewTextFieldMapping()

	bookMapping.AddFieldMappingsAt("Title", indexedText)
	bookMapping.AddFieldMappingsAt("Authors", indexedText)
	bookMapping.AddFieldMappingsAt("Series", indexedText)

	storedText := bleve.NewTextFieldMapping()
	storedText.IncludeInAll = false
	storedText.Index = false
	storedText.SkipFreqNorm = true
	storedText.IncludeTermVectors = false

	bookMapping.AddFieldMappingsAt("LibId", storedText)
	bookMapping.AddFieldMappingsAt("GenresStored", storedText)
	bookMapping.AddFieldMappingsAt("Folder", storedText)
	bookMapping.AddFieldMappingsAt("File", storedText)
	bookMapping.AddFieldMappingsAt("Ext", storedText)
	bookMapping.AddFieldMappingsAt("Archive", storedText)
	bookMapping.AddFieldMappingsAt("Language", storedText)

	storedInt := bleve.NewNumericFieldMapping()
	storedInt.IncludeInAll = false
	storedInt.Index = false
	storedInt.SkipFreqNorm = true
	storedInt.IncludeTermVectors = false

	bookMapping.AddFieldMappingsAt("Size", storedInt)

	indexedInt := bleve.NewNumericFieldMapping()
	indexedInt.IncludeInAll = false
	bookMapping.AddFieldMappingsAt("SeriesNo", indexedInt)

	indexedDate := bleve.NewDateTimeFieldMapping()
	indexedDate.IncludeInAll = false
	bookMapping.AddFieldMappingsAt("PubDate", indexedDate)

	indexMapping := bleve.NewIndexMapping()
	indexMapping.DefaultAnalyzer = analyzer
	indexMapping.AddDocumentMapping("book", bookMapping)
	return indexMapping
}

func getAnalyzer(name string) string {
	analyzers := map[string]string{
		"en": "en",
		"ru": ru.AnalyzerName,
	}

	r, ok := analyzers[name]
	if ok {
		return r
	}
	return "en"
}
