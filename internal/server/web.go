package server

import (
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/vorlif/spreak"

	"github.com/shemanaev/inpxer/internal/config"
	"github.com/shemanaev/inpxer/internal/db"
	"github.com/shemanaev/inpxer/internal/model"
	"github.com/shemanaev/inpxer/ui"
)

type WebHandler struct {
	cfg       *config.MyConfig
	localizer *spreak.Localizer
	indexTpl  *template.Template
	searchTpl *template.Template
}

type pagination struct {
	Last     int
	Page     int
	PrevPage int
	NextPage int
	HasPrev  bool
	HasNext  bool
}

type resultStats struct {
	Total      uint64
	RangeStart int
	RangeEnd   int
}

type arguments struct {
	T          *spreak.Localizer
	Converters []*config.Converter
	Title      string
	Query      string
	Field      string
	Paginator  pagination
	Results    resultStats
	Hits       []*model.Book
}

func NewWebHandler(cfg *config.MyConfig, localizer *spreak.Localizer) *WebHandler {
	indexTpl, err := template.ParseFS(ui.Templates, "templates/index.gohtml", "templates/_*.gohtml")
	if err != nil {
		log.Fatal(err)
	}

	searchTpl, err := template.ParseFS(ui.Templates, "templates/search.gohtml", "templates/_*.gohtml")
	if err != nil {
		log.Fatal(err)
	}

	return &WebHandler{
		cfg:       cfg,
		localizer: localizer,
		indexTpl:  indexTpl,
		searchTpl: searchTpl,
	}
}

func (h *WebHandler) Home(w http.ResponseWriter, r *http.Request) {
	args := arguments{
		T:     h.localizer,
		Title: h.cfg.Title,
	}
	if err := h.indexTpl.Execute(w, args); err != nil {
		internalServerError(w)
	}
}

func (h *WebHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	field := r.URL.Query().Get("field")
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 0
	}

	index, err := db.Open(h.cfg.IndexPath)
	if err != nil {
		log.Printf("Error opening index: %v", err)
		internalServerError(w)
		return
	}
	defer index.Close()

	top, err := index.SearchByField(field, q, page, PageSize)
	if err != nil {
		log.Printf("Error searching: %v", err.Error())
		internalServerError(w)
		return
	}

	totalPages := int(math.Ceil(float64(top.Total) / float64(PageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	paginator := pagination{
		Page: page,
		Last: totalPages - 1,
	}

	if page == 0 {
		paginator.HasPrev = false
	} else {
		paginator.HasPrev = true
		paginator.PrevPage = page - 1
	}

	if page+1 > totalPages-1 {
		paginator.HasNext = false
	} else {
		paginator.HasNext = true
		paginator.NextPage = page + 1
	}

	stats := resultStats{
		Total:      top.Total,
		RangeStart: page*PageSize + 1,
		RangeEnd:   int(math.Min(float64((page+1)*PageSize), float64(top.Total))),
	}

	args := arguments{
		T:          h.localizer,
		Converters: h.cfg.Converters,
		Title:      h.cfg.Title,
		Query:      q,
		Field:      field,
		Paginator:  paginator,
		Results:    stats,
		Hits:       top.Hits,
	}
	if err := h.searchTpl.Execute(w, args); err != nil {
		internalServerError(w)
	}
}
