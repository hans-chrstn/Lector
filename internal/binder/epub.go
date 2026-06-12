package binder

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-shiori/go-epub"
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
	e, err := epub.NewEpub(doc.Title)
	if err != nil {
		return fmt.Errorf("failed to create epub: %v", err)
	}

	e.SetAuthor(doc.Author)

	desc := ""
	if doc.Genres != "" {
		desc += "Genres: " + doc.Genres + "\n"
	}
	if doc.Status != "" {
		desc += "Status: " + doc.Status + "\n"
	}
	if doc.Synopsis != "" {
		desc += "\nSynopsis:\n" + doc.Synopsis
	}
	if desc != "" {
		e.SetDescription(desc)
	}

	var tempFiles []string
	defer func() {
		for _, f := range tempFiles {
			os.Remove(f)
		}
	}()

	if doc.CoverURL != "" {
		coverPath := doc.CoverURL
		isValidLocalPath := false

		if strings.HasPrefix(coverPath, "/uploads/") || strings.HasPrefix(coverPath, "uploads/") {
			coverPath = strings.TrimPrefix(coverPath, "/")
			isValidLocalPath = true
		} else if strings.HasPrefix(coverPath, "http") {
			resp, err := http.Get(coverPath)
			if err == nil && resp.StatusCode == 200 {
				defer resp.Body.Close()
				tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("cover_%d%s", doc.ID, filepath.Ext(coverPath)))
				out, err := os.Create(tmpFile)
				if err == nil {
					io.Copy(out, resp.Body)
					out.Close()
					coverPath = tmpFile
					tempFiles = append(tempFiles, tmpFile)
					isValidLocalPath = true
				}
			}
		}

		if isValidLocalPath {
			if _, err := os.Stat(coverPath); err == nil {
				internalPath, err := e.AddImage(coverPath, "cover"+filepath.Ext(coverPath))
				if err == nil {
					e.SetCover(internalPath, "")
				}
			}
		}
	}

	for _, chapter := range doc.Chapters {
		title := strings.ToLower(strings.TrimSpace(chapter.Title))
		docTitle := strings.ToLower(strings.TrimSpace(doc.Title))

		isTitlePage := title == docTitle || strings.HasPrefix(title, docTitle+":")
		isPreface := title == "preface" || title == "metadata" || title == "introduction" || title == "cover"

		if isTitlePage || isPreface {
			content := strings.ToLower(chapter.Content)
			if strings.Contains(content, "cover image") ||
				strings.Contains(content, "genres:") ||
				strings.Contains(content, "status:") ||
				len(strings.TrimSpace(chapter.Content)) < 300 {
				continue
			}
		}

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
