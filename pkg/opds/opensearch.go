package opds

import "encoding/xml"

type OpenSearchDescription struct {
	XMLName        xml.Name  `xml:"http://a9.com/-/spec/opensearch/1.1/ OpenSearchDescription"`
	ShortName      string    `xml:"ShortName,omitempty"`
	Language       string    `xml:"Language,omitempty"`
	InputEncoding  string    `xml:"InputEncoding,omitempty"`
	OutputEncoding string    `xml:"OutputEncoding,omitempty"`
	Url            SearchUrl `xml:"Url,omitempty"`
}

type SearchUrl struct {
	Type     string `xml:"type,attr"`
	Template string `xml:"template,attr"`
}

func NewOpenSearchDescription(name string, url string) *OpenSearchDescription {
	return &OpenSearchDescription{
		Language:       "*",
		InputEncoding:  "UTF-8",
		OutputEncoding: "UTF-8",
		ShortName:      name,
		Url: SearchUrl{
			Type:     LinkTypeAcquisition,
			Template: url,
		},
	}
}
