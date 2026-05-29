package services

import (
	"archive/zip"
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/nwaples/rardecode/v2"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
)

func processComic(path string) (*models.Document, error) {
	ext := strings.ToLower(filepath.Ext(path))
	var images []string

	if ext == ".cbz" {
		r, err := zip.OpenReader(path)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		for _, f := range r.File {
			if isImage(f.Name) {
				images = append(images, f.Name)
			}
		}
	} else if ext == ".cbr" {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		rr, err := rardecode.NewReader(f)
		if err != nil {
			return nil, err
		}

		for {
			header, err := rr.Next()
			if err != nil {
				break
			}
			if isImage(header.Name) {
				images = append(images, header.Name)
			}
		}
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no images found in archive")
	}

	sort.Slice(images, func(i, j int) bool {
		return naturalLess(images[i], images[j])
	})

	docURL := "local://" + filepath.Base(path)
	var document models.Document
	db.DB.Where("url = ?", docURL).First(&document)
	if document.ID == 0 {
		document = models.Document{URL: docURL}
	}

	document.Title = strings.TrimSuffix(filepath.Base(path), ext)
	document.Source = "local"
	document.IsLocal = true
	document.LocalPath = path
	document.IsInLibrary = true

	if err := db.DB.Save(&document).Error; err != nil {
		return nil, err
	}

	var content strings.Builder
	for _, img := range images {
		src := fmt.Sprintf("/api/documents/%d/archive-image?file=%s", document.ID, url.QueryEscape(img))
		content.WriteString(fmt.Sprintf(`<img src="%s" class="comic-page">`, src))
	}

	chapter := models.Chapter{
		DocumentID: document.ID,
		Title:      "Complete Archive",
		URL:        document.URL + "/complete",
		Content:    content.String(),
		Order:      1,
		Status:     "done",
	}

	db.DB.Unscoped().Where("document_id = ?", document.ID).Delete(&models.Chapter{})
	db.DB.Create(&chapter)

	document.Chapters = []models.Chapter{chapter}
	return &document, nil
}

func GetImageFromArchive(docID uint, fileName string) ([]byte, string, error) {
	var doc models.Document
	if err := db.DB.First(&doc, docID).Error; err != nil {
		return nil, "", err
	}

	ext := strings.ToLower(filepath.Ext(doc.LocalPath))
	if ext == ".cbz" {
		r, err := zip.OpenReader(doc.LocalPath)
		if err != nil {
			return nil, "", err
		}
		defer r.Close()

		for _, f := range r.File {
			if f.Name == fileName {
				rc, err := f.Open()
				if err != nil {
					return nil, "", err
				}
				defer rc.Close()
				data, err := io.ReadAll(rc)
				if err != nil {
					return nil, "", err
				}
				return data, mime.TypeByExtension(filepath.Ext(fileName)), nil
			}
		}
	} else if ext == ".cbr" {
		f, err := os.Open(doc.LocalPath)
		if err != nil {
			return nil, "", err
		}
		defer f.Close()

		rr, err := rardecode.NewReader(f)
		if err != nil {
			return nil, "", err
		}

		for {
			header, err := rr.Next()
			if err != nil {
				break
			}
			if header.Name == fileName {
				data, err := io.ReadAll(rr)
				if err != nil {
					return nil, "", err
				}
				return data, mime.TypeByExtension(filepath.Ext(fileName)), nil
			}
		}
	}

	return nil, "", fmt.Errorf("file not found in archive")
}

func isImage(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" || ext == ".avif" || ext == ".gif"
}

func naturalLess(s1, s2 string) bool {
	i, j := 0, 0
	for i < len(s1) && j < len(s2) {
		c1, c2 := s1[i], s2[j]
		if isDigit(c1) && isDigit(c2) {
			n1, len1 := parseNumber(s1[i:])
			n2, len2 := parseNumber(s2[j:])
			if n1 != n2 {
				return n1 < n2
			}
			i += len1
			j += len2
		} else {
			if c1 != c2 {
				return c1 < c2
			}
			i++
			j++
		}
	}
	return len(s1) < len(s2)
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func parseNumber(s string) (int, int) {
	idx := 0
	for idx < len(s) && isDigit(s[idx]) {
		idx++
	}
	n, _ := strconv.Atoi(s[:idx])
	return n, idx
}
