package opds

import (
	"encoding/xml"
	"time"
)

const (
	ContentType = "application/atom+xml;charset=utf-8"
)

const (
	LinkTypeAcquisition = "application/atom+xml;profile=opds-catalog;kind=acquisition"
	LinkTypeNavigation  = "application/atom+xml;profile=opds-catalog;kind=navigation"
	LinkTypeEntry       = "application/atom+xml;type=entry;profile=opds-catalog"
	LinkTypeOpenSearch  = "application/opensearchdescription+xml"
)

const (
	LinkRelSelf    = "self"
	LinkRelStart   = "start"
	LinkRelFirst   = "first"
	LinkRelLast    = "last"
	LinkRelNext    = "next"
	LinkRelPrev    = "previous"
	LinkRelSearch  = "search"
	LinkRelRelated = "related"

	LinkRelAcquisition = "http://opds-spec.org/acquisition"
	LinkRelImage       = "http://opds-spec.org/image"
	LinkRelThumbnail   = "http://opds-spec.org/image/thumbnail"
)

type Feed struct {
	XMLName       xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
	NamespaceDc   string   `xml:"xmlns:dc,attr"`
	NamespaceOs   string   `xml:"xmlns:os,attr"`
	NamespaceOpds string   `xml:"xmlns:opds,attr"`

	ID      string     `xml:"id"`
	Title   string     `xml:"title"`
	Icon    string     `xml:"icon,omitempty"`
	Link    []Link     `xml:"link"`
	Updated *time.Time `xml:"updated,omitempty"`
	Author  *Author    `xml:"author,omitempty"`
	Entry   []*Entry   `xml:"entry"`

	TotalResults uint64 `xml:"os:totalResults,omitempty"`
	ItemsPerPage int    `xml:"os:itemsPerPage,omitempty"`
}

type Entry struct {
	ID        string     `xml:"id"`
	Title     string     `xml:"title"`
	Link      []Link     `xml:"link"`
	Published *time.Time `xml:"published"`
	Updated   *time.Time `xml:"updated"`
	Summary   *Text      `xml:"summary"`
	Content   *Text      `xml:"content"`
	Category  []Category `xml:"category"`
	Author    []Author   `xml:"author"`

	Issued     *time.Time `xml:"dc:issued"`
	Language   string     `xml:"dc:language,omitempty"`
	Publisher  string     `xml:"dc:publisher,omitempty"`
	Identifier string     `xml:"dc:identifier,omitempty"`
}

type Author struct {
	Name string `xml:"name"`
	Uri  string `xml:"uri,omitempty"`
}

type Category struct {
	Term   string `xml:"term,attr"`
	Label  string `xml:"label,attr"`
	Scheme string `xml:"scheme,attr,omitempty"`
}

type Link struct {
	Rel      string `xml:"rel,attr,omitempty"`
	Type     string `xml:"type,attr,omitempty"`
	Href     string `xml:"href,attr"`
	HrefLang string `xml:"hreflang,attr,omitempty"`
	Title    string `xml:"title,attr,omitempty"`
	Length   uint   `xml:"length,attr,omitempty"`
}

type Text struct {
	Type string `xml:"type,attr"`
	Body string `xml:",chardata"`
}

func NewFeed() *Feed {
	return &Feed{
		NamespaceDc:   "http://purl.org/dc/terms/",
		NamespaceOs:   "http://a9.com/-/spec/opensearch/1.1/",
		NamespaceOpds: "http://opds-spec.org/2010/catalog",
	}
}

func NewText(s string) *Text {
	return &Text{
		Type: "text",
		Body: s,
	}
}
