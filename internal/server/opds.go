package server

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"math"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/vorlif/spreak"

	"github.com/shemanaev/inpxer/internal/config"
	"github.com/shemanaev/inpxer/internal/db"
	"github.com/shemanaev/inpxer/internal/i18n"
	"github.com/shemanaev/inpxer/internal/model"
	"github.com/shemanaev/inpxer/pkg/opds"
)

type OpdsHandler struct {
	cfg *config.MyConfig
	t   *spreak.Localizer
}

func NewOpdsHandler(cfg *config.MyConfig, localizer *spreak.Localizer) *OpdsHandler {
	return &OpdsHandler{
		cfg: cfg,
		t:   localizer,
	}
}

func (h *OpdsHandler) OpenSearchDescription(w http.ResponseWriter, r *http.Request) {
	templateUrl, err := url.JoinPath(h.cfg.FullUrl, "/opds/search")
	if err != nil {
		log.Fatalf("Error combining url: %v", err)
	}

	description := opds.NewOpenSearchDescription(h.cfg.Title, templateUrl+"?q={searchTerms}&page={startPage?}")

	content, _ := xml.MarshalIndent(description, "  ", "    ")
	w.Header().Add("Content-Type", opds.ContentType)
	http.ServeContent(w, r, "opensearch.xml", time.Now(), bytes.NewReader(content))
}

func (h *OpdsHandler) Root(w http.ResponseWriter, r *http.Request) {
	index, err := db.Open(h.cfg.IndexPath, h.cfg.Storage)
	if err != nil {
		log.Printf("Error opening index: %v", err)
		internalServerError(w)
		return
	}
	defer index.Close()

	books, err := index.GetMostRecentBooks(PageSize)
	if err != nil {
		log.Printf("Error retrieving recent books: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal server error")
		return
	}

	entries := h.makeBooksList(books)

	h.serveFeed(w, r, "root", entries, nil, uint64(len(books)))
}

func (h *OpdsHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	field := r.URL.Query().Get("field")
	if field == "" {
		field = "_all"
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 0
	}

	index, err := db.Open(h.cfg.IndexPath, h.cfg.Storage)
	if err != nil {
		log.Printf("Error opening index: %v", err)
		internalServerError(w)
		return
	}
	defer index.Close()

	top, err := index.SearchByField(field, q, page, PageSize)
	if err != nil {
		log.Printf("Error searching: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal server error")
		return
	}

	entries := h.makeBooksList(top.Hits)
	totalPages := int(math.Ceil(float64(top.Total) / float64(PageSize)))

	var links []opds.Link
	link := fmt.Sprintf("/opds/search?q=%s", url.QueryEscape(q))
	links = append(links, opds.Link{
		Rel:  opds.LinkRelFirst,
		Type: opds.LinkTypeNavigation,
		Href: link,
	})

	if page > 0 {
		link := fmt.Sprintf("/opds/search?q=%s&page=%d", url.QueryEscape(q), page-1)
		links = append(links, opds.Link{
			Rel:  opds.LinkRelPrev,
			Type: opds.LinkTypeNavigation,
			Href: link,
		})
	}

	if page+1 <= totalPages-1 {
		link := fmt.Sprintf("/opds/search?q=%s&page=%d", url.QueryEscape(q), page+1)
		links = append(links, opds.Link{
			Rel:  opds.LinkRelNext,
			Type: opds.LinkTypeNavigation,
			Href: link,
		})
	}

	if totalPages > 1 && page != totalPages-1 {
		link := fmt.Sprintf("/opds/search?q=%s&page=%d", url.QueryEscape(q), totalPages-1)
		links = append(links, opds.Link{
			Rel:  opds.LinkRelLast,
			Type: opds.LinkTypeNavigation,
			Href: link,
		})
	}

	h.serveFeed(w, r, "search", entries, links, top.Total)
}

func (h *OpdsHandler) serveFeed(w http.ResponseWriter, r *http.Request, id string, entries []*opds.Entry, links []opds.Link, totalResults uint64) {
	now := time.Now()
	feed := opds.NewFeed()
	feed.ID = id
	feed.Title = h.cfg.Title
	feed.Updated = &now
	feed.Entry = entries
	feed.ItemsPerPage = PageSize
	feed.TotalResults = totalResults

	navLinks := []opds.Link{
		{
			Rel:  opds.LinkRelStart,
			Type: opds.LinkTypeNavigation,
			Href: "/opds",
		},
		{
			Rel:  opds.LinkRelSearch,
			Type: opds.LinkTypeOpenSearch,
			Href: "/opensearch.xml",
		},
		{
			Rel:  opds.LinkRelSearch,
			Type: opds.ContentType,
			Href: "/opds/search?q={searchTerms}",
		},
	}

	feed.Link = append(feed.Link, navLinks...)
	feed.Link = append(feed.Link, links...)

	content, _ := xml.MarshalIndent(feed, "  ", "    ")
	w.Header().Add("Content-Type", opds.ContentType)
	content = append([]byte(xml.Header), content...)
	http.ServeContent(w, r, "feed.xml", time.Now(), bytes.NewReader(content))
}

func (h *OpdsHandler) makeBooksList(books []*model.Book) []*opds.Entry {
	entries := make([]*opds.Entry, 0)
	for _, book := range books {
		entry := &opds.Entry{
			ID:       fmt.Sprintf("book:%s", book.LibId),
			Title:    book.CleanTitle(),
			Issued:   &book.PubDate,
			Language: book.Language,
		}
		entry.Content = opds.NewText(h.t.Getf("Original title: %s", book.Title))

		for _, author := range book.Authors {
			entry.Author = append(entry.Author, opds.Author{Name: author.FormattedName(h.cfg.AuthorNameFormat)})
			entry.Link = append(entry.Link, opds.Link{
				Rel:   opds.LinkRelRelated,
				Type:  opds.LinkTypeNavigation,
				Href:  "/opds/search?q=" + url.QueryEscape(author.String()),
				Title: h.t.Getf("Search books by %s", author.FormattedName(h.cfg.AuthorNameFormat)),
			})
		}

		for _, genre := range book.Genres {
			cat := h.t.DGet(i18n.GenresDomain, genre)
			entry.Category = append(entry.Category, opds.Category{
				Term:  cat,
				Label: cat,
			})
		}

		var fileMime string
		if strings.HasSuffix(book.File.Name, ".fb2.zip") {
			fileMime = mime.TypeByExtension(".fb2.zip")
		} else {
			fileMime = mime.TypeByExtension("." + book.File.Ext)
		}

		entry.Link = append(entry.Link, opds.Link{
			Rel:  opds.LinkRelAcquisition,
			Type: fileMime,
			Href: fmt.Sprintf("/download/%s", book.LibId),
		})

		for _, converter := range h.cfg.Converters {
			if strings.EqualFold(converter.From, book.File.Ext) {
				entry.Link = append(entry.Link, opds.Link{
					Rel:  opds.LinkRelAcquisition,
					Type: mime.TypeByExtension("." + converter.To),
					Href: fmt.Sprintf("/download/%s/%s", book.LibId, converter.To),
				})
			}
		}

		entries = append(entries, entry)
	}

	return entries
}
