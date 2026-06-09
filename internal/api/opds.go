package api

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
)

type OPDSFeed struct {
	XMLName xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	ID      string      `xml:"id"`
	Title   string      `xml:"title"`
	Updated string      `xml:"updated"`
	Author  OPDSAuthor  `xml:"author"`
	Links   []OPDSLink  `xml:"link"`
	Entries []OPDSEntry `xml:"entry"`
}

type OPDSAuthor struct {
	Name string `xml:"name"`
}

type OPDSLink struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Href string `xml:"href,attr"`
	Type string `xml:"type,attr,omitempty"`
}

type OPDSEntry struct {
	ID      string     `xml:"id"`
	Title   string     `xml:"title"`
	Author  OPDSAuthor `xml:"author"`
	Updated string     `xml:"updated"`
	Summary string     `xml:"summary,omitempty"`
	Links   []OPDSLink `xml:"link"`
}

func (h *API) GetOPDSRoot(c *fiber.Ctx) error {
	baseURL := c.Protocol() + "://" + c.Hostname()
	if c.Port() != "" && c.Port() != "80" && c.Port() != "443" {
		baseURL = baseURL + ":" + c.Port()
	}

	feed := OPDSFeed{
		ID:      "urn:lector:root",
		Title:   "Lector Library",
		Updated: time.Now().Format(time.RFC3339),
		Author:  OPDSAuthor{Name: "Lector"},
		Links: []OPDSLink{
			{Rel: "self", Href: "/api/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
			{Rel: "start", Href: "/api/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
		Entries: []OPDSEntry{
			{
				ID:      "urn:lector:all",
				Title:   "All Documents",
				Updated: time.Now().Format(time.RFC3339),
				Links: []OPDSLink{
					{Rel: "subsection", Href: "/api/opds/all", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
				},
			},
		},
	}

	xmlData, _ := xml.MarshalIndent(feed, "", "  ")
	c.Set("Content-Type", "application/atom+xml; charset=utf-8")
	return c.Send([]byte(xml.Header + string(xmlData)))
}

func (h *API) GetOPDSAll(c *fiber.Ctx) error {
	baseURL := c.Protocol() + "://" + c.Hostname()
	if c.Port() != "" && c.Port() != "80" && c.Port() != "443" {
		baseURL = baseURL + ":" + c.Port()
	}

	var documents []models.Document
	db.DB.WithContext(c.UserContext()).Find(&documents)

	feed := OPDSFeed{
		ID:      "urn:lector:all",
		Title:   "All Documents",
		Updated: time.Now().Format(time.RFC3339),
		Author:  OPDSAuthor{Name: "Lector"},
		Links: []OPDSLink{
			{Rel: "self", Href: "/api/opds/all", Type: "application/atom+xml;profile=opds-catalog;kind=acquisition"},
			{Rel: "start", Href: "/api/opds", Type: "application/atom+xml;profile=opds-catalog;kind=navigation"},
		},
	}

	for _, doc := range documents {
		entry := OPDSEntry{
			ID:      fmt.Sprintf("urn:lector:doc:%d", doc.ID),
			Title:   doc.Title,
			Author:  OPDSAuthor{Name: doc.Author},
			Updated: doc.UpdatedAt.Format(time.RFC3339),
			Summary: doc.Synopsis,
		}

		if doc.CoverURL != "" {
			coverURL := doc.CoverURL
			if !strings.HasPrefix(coverURL, "http") {
				coverURL = baseURL + "/api/proxy-image?url=" + url.QueryEscape(coverURL)
			}
			entry.Links = append(entry.Links, OPDSLink{
				Rel:  "http://opds-spec.org/image",
				Href: coverURL,
				Type: "image/jpeg",
			})
			entry.Links = append(entry.Links, OPDSLink{
				Rel:  "http://opds-spec.org/image/thumbnail",
				Href: coverURL,
				Type: "image/jpeg",
			})
		}

		entry.Links = append(entry.Links, OPDSLink{
			Rel:  "http://opds-spec.org/acquisition",
			Href: fmt.Sprintf("%s/api/documents/%d/export?format=epub", baseURL, doc.ID),
			Type: "application/epub+zip",
		})

		feed.Entries = append(feed.Entries, entry)
	}

	xmlData, _ := xml.MarshalIndent(feed, "", "  ")
	c.Set("Content-Type", "application/atom+xml; charset=utf-8")
	return c.Send([]byte(xml.Header + string(xmlData)))
}
