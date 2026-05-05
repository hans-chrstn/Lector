package binder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmaupin/go-epub"
	"github.com/user/lector/internal/models"
)

type Binder interface {
	Bind(doc models.Document, outputPath string) error
}

type EPUBBinder struct{}

func NewEPUBBinder() *EPUBBinder {
	return &EPUBBinder{}
}

func BindEPUB(doc *models.Document, outputPath string) error {
	b := NewEPUBBinder()
	return b.Bind(*doc, outputPath)
}

func (b *EPUBBinder) Bind(doc models.Document, outputPath string) error {
	e := epub.NewEpub(doc.Title)
	e.SetAuthor(doc.Author)

	for _, chapter := range doc.Chapters {
		sectionContent := fmt.Sprintf("<h1>%s</h1>%s", chapter.Title, chapter.Content)
		_, err := e.AddSection(sectionContent, chapter.Title, "", "")
		if err != nil {
			return fmt.Errorf("failed to add chapter %s: %v", chapter.Title, err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	return e.Write(outputPath)
}
