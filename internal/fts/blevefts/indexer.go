package blevefts

import (
	"log"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/lang/ru"
	"github.com/blevesearch/bleve/v2/index/scorch"
	"github.com/blevesearch/bleve/v2/index/upsidedown/store/boltdb"
	"github.com/blevesearch/bleve/v2/mapping"

	"github.com/shemanaev/inpxer/internal/fts"
)

type Indexer struct {
	fts.Indexer
	index bleve.Index
}

func Open(path string) (*Indexer, error) {
	idx, err := bleve.Open(path)
	if err != nil {
		return nil, err
	}

	res := Indexer{
		index: idx,
	}
	return &res, nil
}

func Create(path, language string) (*Indexer, error) {
	idx, err := bleve.Open(path)
	if err != nil {
		analyzer := getAnalyzer(language)
		idx, err = bleve.NewUsing(path, createBookMapping(analyzer), scorch.Name, boltdb.Name, nil)
		if err != nil {
			return nil, err
		}
	}

	res := Indexer{
		index: idx,
	}
	return &res, nil
}

func (i *Indexer) Close() error {
	return i.index.Close()
}

func (i *Indexer) AddBooks(books []*fts.Book) error {
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

func (i *Indexer) SearchByField(field, s string, page, pageSize int) (*fts.SearchResult, error) {
	query := bleve.NewMatchQuery(s)
	query.SetField(field)
	search := bleve.NewSearchRequestOptions(query, pageSize, page*pageSize, false)

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

	var hitIds []string
	for _, v := range searchResults.Hits {
		hitIds = append(hitIds, v.ID)
	}

	res := fts.SearchResult{
		Total: searchResults.Total,
		Hits:  hitIds,
	}
	return &res, nil
}

func (i *Indexer) GetMostRecentBooks(count int) ([]string, error) {
	t := time.Now()
	p := t.AddDate(-1, 0, 0)
	query := bleve.NewDateRangeQuery(p, t)
	query.SetField("PubDate")
	search := bleve.NewSearchRequestOptions(query, count, 0, false)
	search.Fields = []string{}
	search.SortBy([]string{"-PubDate"})

	searchResults, err := i.index.Search(search)
	if err != nil {
		return nil, err
	}

	var hitIds []string
	for _, v := range searchResults.Hits {
		hitIds = append(hitIds, v.ID)
	}

	return hitIds, nil
}

func createBookMapping(analyzer string) *mapping.IndexMappingImpl {
	bookMapping := bleve.NewDocumentMapping()

	indexedText := bleve.NewTextFieldMapping()
	indexedText.Store = false

	bookMapping.AddFieldMappingsAt("Title", indexedText)
	bookMapping.AddFieldMappingsAt("Authors", indexedText)
	bookMapping.AddFieldMappingsAt("Series", indexedText)

	disabled := bleve.NewDocumentDisabledMapping()
	bookMapping.AddSubDocumentMapping("LibId", disabled)

	indexedInt := bleve.NewNumericFieldMapping()
	indexedInt.Store = false
	indexedInt.IncludeInAll = false
	bookMapping.AddFieldMappingsAt("SeriesNo", indexedInt)

	indexedDate := bleve.NewDateTimeFieldMapping()
	indexedDate.Store = false
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
