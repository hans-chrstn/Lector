package services

import (
	"encoding/xml"
	"os"
	"strings"

	"github.com/user/lector/internal/models"
)

type OPFMetadata struct {
	XMLName  xml.Name `xml:"package"`
	Metadata struct {
		Title       string   `xml:"http://purl.org/dc/elements/1.1/ title"`
		Creator     string   `xml:"http://purl.org/dc/elements/1.1/ creator"`
		Description string   `xml:"http://purl.org/dc/elements/1.1/ description"`
		Subject     []string `xml:"http://purl.org/dc/elements/1.1/ subject"`
		Meta        []struct {
			Name    string `xml:"name,attr"`
			Content string `xml:"content,attr"`
		} `xml:"meta"`
	} `xml:"metadata"`
}

func ParseSidecarOPF(path string) (*models.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var opf OPFMetadata
	if err := xml.Unmarshal(data, &opf); err != nil {
		return nil, err
	}

	doc := &models.Document{
		Title:    opf.Metadata.Title,
		Author:   opf.Metadata.Creator,
		Synopsis: opf.Metadata.Description,
		Genres:   strings.Join(opf.Metadata.Subject, ", "),
	}

	for _, m := range opf.Metadata.Meta {
		if m.Name == "status" {
			doc.Status = strings.ToLower(strings.TrimSpace(m.Content))
		}
	}

	return doc, nil
}
