package reader

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const maxEntry = 8 << 20
const maxEntries = 3000

type Chapter struct {
	Title string `json:"title"`
	Text  string `json:"text,omitempty"`
	Image int    `json:"image,omitempty"`
}
type Content struct {
	Chapters []Chapter `json:"chapters"`
}

func openArchive(file string) (*zip.ReadCloser, error) {
	z, err := zip.OpenReader(file)
	if err != nil {
		return nil, errors.New("invalid document archive")
	}
	if len(z.File) > maxEntries {
		_ = z.Close()
		return nil, errors.New("document contains too many entries")
	}
	var size uint64
	for _, f := range z.File {
		size += f.UncompressedSize64
		if size > 256<<20 {
			_ = z.Close()
			return nil, errors.New("expanded document exceeds 256 MB")
		}
	}
	return z, nil
}
func readEntry(f *zip.File) ([]byte, error) {
	if f.UncompressedSize64 > maxEntry {
		return nil, errors.New("document entry exceeds 8 MB")
	}
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	b, err := io.ReadAll(io.LimitReader(r, maxEntry+1))
	if len(b) > maxEntry {
		return nil, errors.New("document entry exceeds 8 MB")
	}
	return b, err
}
func comicFiles(z *zip.ReadCloser) []*zip.File {
	files := []*zip.File{}
	for _, f := range z.File {
		switch strings.ToLower(path.Ext(f.Name)) {
		case ".jpg", ".jpeg", ".png", ".webp", ".gif":
			files = append(files, f)
		}
	}
	slices.SortFunc(files, func(a, b *zip.File) int { return strings.Compare(a.Name, b.Name) })
	return files
}
func Comic(file string) (Content, error) {
	z, err := openArchive(file)
	if err != nil {
		return Content{}, err
	}
	defer z.Close()
	result := Content{Chapters: []Chapter{}}
	for i := range comicFiles(z) {
		result.Chapters = append(result.Chapters, Chapter{Title: fmt.Sprintf("Page %d", i+1), Image: i + 1})
	}
	if len(result.Chapters) == 0 {
		return result, errors.New("no supported images in document")
	}
	return result, nil
}
func ComicImage(file string, index int) ([]byte, error) {
	z, err := openArchive(file)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	files := comicFiles(z)
	if index < 1 || index > len(files) {
		return nil, errors.New("page not found")
	}
	return readEntry(files[index-1])
}
func EPUB(file string) (Content, error) {
	z, err := openArchive(file)
	if err != nil {
		return Content{}, err
	}
	defer z.Close()
	files := map[string]*zip.File{}
	for _, f := range z.File {
		files[path.Clean(f.Name)] = f
	}
	read := func(name string) ([]byte, error) {
		f, ok := files[path.Clean(name)]
		if !ok {
			return nil, errors.New("incomplete EPUB document")
		}
		return readEntry(f)
	}
	b, err := read("META-INF/container.xml")
	if err != nil {
		return Content{}, err
	}
	var container struct {
		Files []struct {
			Path string `xml:"full-path,attr"`
		} `xml:"rootfiles>rootfile"`
	}
	if err = xml.Unmarshal(b, &container); err != nil || len(container.Files) == 0 {
		return Content{}, errors.New("invalid EPUB container")
	}
	opf := container.Files[0].Path
	b, err = read(opf)
	if err != nil {
		return Content{}, err
	}
	var book struct {
		Manifest []struct {
			ID   string `xml:"id,attr"`
			Href string `xml:"href,attr"`
		} `xml:"manifest>item"`
		Spine []struct {
			ID string `xml:"idref,attr"`
		} `xml:"spine>itemref"`
	}
	if err = xml.Unmarshal(b, &book); err != nil {
		return Content{}, err
	}
	manifest := map[string]string{}
	for _, m := range book.Manifest {
		manifest[m.ID] = path.Join(path.Dir(opf), m.Href)
	}
	result := Content{Chapters: []Chapter{}}
	for idx, s := range book.Spine {
		b, err := read(manifest[s.ID])
		if err != nil {
			return result, err
		}
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(b)))
		if err != nil {
			return result, err
		}
		doc.Find("script,style,iframe,object,svg").Remove()
		title := strings.TrimSpace(doc.Find("h1,h2,title").First().Text())
		if title == "" {
			title = fmt.Sprintf("Chapter %d", idx+1)
		}
		doc.Find("br").ReplaceWithHtml("\n")
		doc.Find("p,div,h1,h2,h3,li").Each(func(_ int, s *goquery.Selection) { s.AppendHtml("\n\n") })
		text := strings.TrimSpace(doc.Find("body").Text())
		if text == "" {
			text = strings.TrimSpace(doc.Text())
		}
		result.Chapters = append(result.Chapters, Chapter{Title: title, Text: text})
	}
	if len(result.Chapters) == 0 {
		return result, errors.New("EPUB has no chapters")
	}
	return result, nil
}
