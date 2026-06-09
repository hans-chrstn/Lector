package services

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/lector/internal/core/sanitizer"
	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
)

type EPUBPackage struct {
	XMLName  xml.Name `xml:"package"`
	Metadata struct {
		Title       string   `xml:"title"`
		Creator     string   `xml:"creator"`
		Description string   `xml:"description"`
		Subject     []string `xml:"subject"`
		Meta        []struct {
			Name     string `xml:"name,attr"`
			Property string `xml:"property,attr"`
			Content  string `xml:"content,attr"`
			Value    string `xml:",chardata"`
		} `xml:"meta"`
	} `xml:"metadata"`
	Manifest struct {
		Items []struct {
			ID         string `xml:"id,attr"`
			Href       string `xml:"href,attr"`
			MediaType  string `xml:"media-type,attr"`
			Properties string `xml:"properties,attr"`
		} `xml:"item"`
	} `xml:"manifest"`
	Spine struct {
		ItemRefs []struct {
			IDRef string `xml:"idref,attr"`
		} `xml:"itemref"`
	} `xml:"spine"`
}

type EPUBFixer struct {
	Files map[string]bool
}

func NewEPUBFixer(r *zip.ReadCloser) *EPUBFixer {
	files := make(map[string]bool)
	for _, f := range r.File {
		files[filepath.ToSlash(f.Name)] = true
	}
	return &EPUBFixer{Files: files}
}

func (f *EPUBFixer) FixChapter(html string, currentPath string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return html
	}
	dir := filepath.Dir(currentPath)
	doc.Find("img").Each(func(i int, sel *goquery.Selection) {
		src, _ := sel.Attr("src")
		if src == "" {
			sel.Remove()
			return
		}
		absSrc := f.resolvePath(dir, src)
		if !f.Files[absSrc] {
			sel.Remove()
		}
	})
	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href, _ := sel.Attr("href")
		if href == "" || strings.HasPrefix(href, "http") || strings.HasPrefix(href, "mailto") {
			return
		}
		linkPath := strings.Split(href, "#")[0]
		if linkPath == "" {
			return
		}
		absHref := f.resolvePath(dir, linkPath)
		if !f.Files[absHref] {
			sel.ReplaceWithSelection(sel.Contents())
		}
	})
	h, _ := doc.Find("body").Html()
	if h == "" {
		h, _ = doc.Html()
	}
	return h
}

func (f *EPUBFixer) resolvePath(baseDir, relPath string) string {
	if filepath.IsAbs(relPath) {
		return strings.TrimPrefix(relPath, "/")
	}
	return filepath.ToSlash(filepath.Join(baseDir, relPath))
}

func processEPUB(path string) (*models.Document, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %v", err)
	}
	defer r.Close()

	for _, f := range r.File {
		fPath := filepath.ToSlash(f.Name)
		if strings.Contains(fPath, "..") || strings.HasPrefix(fPath, "/") {
			return nil, fmt.Errorf("security: invalid file path in epub: %s", fPath)
		}
	}

	opfPath := ""
	for _, f := range r.File {
		if f.Name == "META-INF/container.xml" {
			rc, _ := f.Open()
			limitRC := io.LimitReader(rc, 1*1024*1024)
			content, _ := io.ReadAll(limitRC)
			rc.Close()
			xmlStr := string(content)
			opfPath = stringBetween(xmlStr, "full-path=\"", "\"")
			if opfPath == "" {
				opfPath = stringBetween(xmlStr, "full-path='", "'")
			}
			break
		}
	}

	if opfPath == "" {
		return nil, fmt.Errorf("invalid epub: missing container.xml or opf path")
	}

	var opfContent []byte
	for _, f := range r.File {
		if f.Name == opfPath {
			rc, _ := f.Open()
			limitRC := io.LimitReader(rc, 10*1024*1024)
			opfContent, _ = io.ReadAll(limitRC)
			rc.Close()
			break
		}
	}

	var pkg EPUBPackage
	if err := xml.Unmarshal(opfContent, &pkg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OPF: %v", err)
	}

	title := pkg.Metadata.Title
	if title == "" {
		title = filepath.Base(path)
	}
	author := pkg.Metadata.Creator
	docURL := "local://" + filepath.Base(path)

	var document models.Document
	db.DB.Where("url = ?", docURL).First(&document)
	if document.ID == 0 {
		document = models.Document{URL: docURL}
	}

	document.Title = title
	document.Author = author
	document.Genres = strings.Join(pkg.Metadata.Subject, ", ")
	document.Source = "local"
	document.IsLocal = true
	document.LocalPath = path
	document.IsInLibrary = true

	rawDesc := pkg.Metadata.Description
	if rawDesc != "" && rawDesc != "nil" {
		lines := strings.Split(rawDesc, "\n")
		foundMeta := false
		var cleanSynopsis []string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "Genres:") {
				document.Genres = strings.TrimSpace(strings.TrimPrefix(line, "Genres:"))
				foundMeta = true
			} else if strings.HasPrefix(line, "Status:") {
				document.Status = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(line, "Status:")))
				foundMeta = true
			} else if strings.HasPrefix(line, "Synopsis:") {
				foundMeta = true
				continue
			} else {
				cleanSynopsis = append(cleanSynopsis, line)
			}
		}

		if foundMeta {
			document.Synopsis = strings.Join(cleanSynopsis, "\n")
		} else {
			document.Synopsis = rawDesc
		}
	}

	for _, m := range pkg.Metadata.Meta {
		if m.Name == "status" || m.Property == "status" {
			val := m.Content
			if val == "" {
				val = m.Value
			}
			if val != "" {
				document.Status = strings.ToLower(strings.TrimSpace(val))
			}
		}
	}

	coverHref := ""
	for _, item := range pkg.Manifest.Items {
		if strings.Contains(item.Properties, "cover-image") {
			coverHref = item.Href
			break
		}
	}

	if coverHref == "" {
		for _, m := range pkg.Metadata.Meta {
			if m.Name == "cover" {
				for _, item := range pkg.Manifest.Items {
					if item.ID == m.Content {
						coverHref = item.Href
						break
					}
				}
			}
		}
	}

	if coverHref != "" {
		opfDir := filepath.Dir(opfPath)
		fullCoverPath := filepath.ToSlash(filepath.Join(opfDir, coverHref))
		for _, f := range r.File {
			if filepath.ToSlash(f.Name) == fullCoverPath {
				rc, _ := f.Open()
				limitRC := io.LimitReader(rc, 20*1024*1024)
				localCoverName := fmt.Sprintf("cover_%s%s", strings.ReplaceAll(filepath.Base(path), " ", "_"), filepath.Ext(f.Name))
				localCoverPath := filepath.Join("uploads", localCoverName)
				out, _ := os.Create(localCoverPath)
				io.Copy(out, limitRC)
				out.Close()
				rc.Close()
				document.CoverURL = "/uploads/" + localCoverName
				break
			}
		}
	}

	manifest := make(map[string]string)
	for _, item := range pkg.Manifest.Items {
		manifest[item.ID] = item.Href
	}

	var chapters []models.Chapter
	fixer := NewEPUBFixer(r)
	opfDir := filepath.Dir(opfPath)
	order := 1
	for _, ref := range pkg.Spine.ItemRefs {
		relHref, ok := manifest[ref.IDRef]
		if !ok {
			continue
		}
		fullHref := filepath.ToSlash(filepath.Join(opfDir, relHref))
		var targetFile *zip.File
		for _, f := range r.File {
			if filepath.ToSlash(f.Name) == fullHref {
				targetFile = f
				break
			}
		}
		if targetFile == nil {
			continue
		}
		rc, _ := targetFile.Open()
		limitRC := io.LimitReader(rc, 50*1024*1024)
		content, _ := io.ReadAll(limitRC)
		rc.Close()
		cDoc, _ := goquery.NewDocumentFromReader(bytes.NewReader(content))
		chTitle := strings.TrimSpace(cDoc.Find("title").First().Text())
		if chTitle == "" || strings.EqualFold(chTitle, "unknown") || strings.EqualFold(chTitle, "chapter") {
			chTitle = ""
			cDoc.Find("h1, h2, h3").Each(func(i int, s *goquery.Selection) {
				if chTitle == "" {
					chTitle = strings.TrimSpace(s.Text())
				}
			})
		}
		if chTitle == "" {
			chTitle = fmt.Sprintf("Chapter %d", order)
		}
		body, _ := cDoc.Find("body").Html()
		if body == "" {
			body = string(content)
		}

		lowerTitle := strings.ToLower(chTitle)
		lowerDocTitle := strings.ToLower(document.Title)
		isTitlePage := lowerTitle == lowerDocTitle || strings.HasPrefix(lowerTitle, lowerDocTitle+":")
		isPreface := lowerTitle == "preface" || lowerTitle == "metadata" || lowerTitle == "introduction" || lowerTitle == "cover"

		if isTitlePage || isPreface {
			if document.Synopsis == "" || document.Synopsis == "nil" || len(document.Synopsis) < 10 {
				txt := strings.TrimSpace(cDoc.Find("div, p").Text())
				if len(txt) > 20 {
					document.Synopsis = txt
				}
			}
			if document.Genres == "" || document.Genres == "nil" {
				cDoc.Find("p").Each(func(i int, s *goquery.Selection) {
					t := s.Text()
					if strings.Contains(t, "Genres:") {
						document.Genres = strings.TrimSpace(strings.ReplaceAll(t, "Genres:", ""))
					}
					if strings.Contains(t, "Status:") {
						document.Status = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(t, "Status:", "")))
					}
				})
			}

			cleanBody := strings.ToLower(cDoc.Text())
			shouldSkip := strings.Contains(cleanBody, "cover image") ||
				strings.Contains(cleanBody, "genres:") ||
				strings.Contains(cleanBody, "status:") ||
				len(strings.TrimSpace(cleanBody)) < 400

			if shouldSkip {
				continue
			}
		}

		fixedBody := fixer.FixChapter(body, fullHref)
		cleanBody := sanitizer.CleanHTML(fixedBody, chTitle)

		chapters = append(chapters, models.Chapter{
			DocumentID: document.ID,
			Title:      chTitle,
			URL:        document.URL + "/" + fullHref,
			Content:    cleanBody,
			Order:      order,
			Status:     "done",
		})
		order++
	}

	remoteMeta := FetchRemoteMetadata(document.Title, document.Author)
	if remoteMeta != nil {
		if document.CoverURL == "" && remoteMeta.CoverURL != "" {
			resp, err := http.Get(remoteMeta.CoverURL)
			if err == nil && resp.StatusCode == 200 {
				localCoverName := fmt.Sprintf("cover_%s.jpg", strings.ReplaceAll(document.Title, " ", "_"))
				localCoverPath := filepath.Join("uploads", localCoverName)
				out, _ := os.Create(localCoverPath)
				io.Copy(out, resp.Body)
				out.Close()
				resp.Body.Close()
				document.CoverURL = "/uploads/" + localCoverName
			}
		}
		if (document.Synopsis == "" || strings.EqualFold(document.Synopsis, "no description") || document.Synopsis == "nil") && remoteMeta.Synopsis != "" {
			document.Synopsis = remoteMeta.Synopsis
		}
	}

	if err := db.DB.Save(&document).Error; err != nil {
		return nil, err
	}

	db.DB.Unscoped().Where("document_id = ?", document.ID).Delete(&models.Chapter{})
	db.DB.CreateInBatches(chapters, 100)
	document.Chapters = chapters
	return &document, nil
}
