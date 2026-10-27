package httpapi

import (
	"crypto/rand"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/user/lector/internal/library"
	"github.com/user/lector/internal/reader"
)

func (s *Server) importFile(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, s.config.MaxUpload+(1<<20))
	multipart, err := r.MultipartReader()
	if err != nil {
		problem(w, 400, "Choose a file to import")
		return
	}
	part, err := multipart.NextPart()
	if err != nil || part.FormName() != "file" {
		problem(w, 400, "Expected a file field")
		return
	}
	defer func() { _ = part.Close() }()
	name := filepath.Base(part.FileName())
	ext := strings.ToLower(filepath.Ext(name))
	kind := "book"
	switch ext {
	case ".epub", ".pdf", ".txt":
	case ".cbz":
		kind = "manga"
	case ".mp4", ".webm":
		kind = "video"
	default:
		problem(w, 400, "Supported files: EPUB, PDF, TXT, CBZ, MP4, WebM")
		return
	}
	if err = os.MkdirAll(s.config.Files, 0700); err != nil {
		s.fail(w, err)
		return
	}
	filename := rand.Text() + ext
	path := filepath.Join(s.config.Files, filename)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		s.fail(w, err)
		return
	}
	keep := false
	defer func() {
		if !keep {
			_ = os.Remove(path)
		}
	}()
	size, copyErr := io.Copy(f, io.LimitReader(part, s.config.MaxUpload+1))
	closeErr := f.Close()
	if copyErr != nil || size > s.config.MaxUpload {
		problem(w, 413, "File exceeds the upload limit")
		return
	}
	if closeErr != nil {
		s.fail(w, closeErr)
		return
	}
	if size == 0 {
		problem(w, 400, "The file is empty")
		return
	}
	if _, err = multipart.NextPart(); err != io.EOF {
		problem(w, 400, "Import one file at a time")
		return
	}
	total := 0
	switch ext {
	case ".epub":
		var content reader.Content
		content, err = reader.EPUB(path)
		total = len(content.Chapters)
	case ".cbz":
		var content reader.Content
		content, err = reader.Comic(path)
		total = len(content.Chapters)
	case ".txt":
		if size > 8<<20 {
			err = errors.New("text files must be 8 MB or smaller")
		} else {
			var b []byte
			b, err = os.ReadFile(path)
			if err == nil && !utf8.Valid(b) {
				err = errors.New("text files must use UTF-8")
			}
			total = 1
		}
	case ".pdf":
		var f *os.File
		f, err = os.Open(path)
		if err == nil {
			var magic [5]byte
			_, err = io.ReadFull(f, magic[:])
			_ = f.Close()
			if string(magic[:]) != "%PDF-" {
				err = errors.New("invalid PDF document")
			}
		}
	}
	if err != nil {
		problem(w, 400, err.Error())
		return
	}
	title := strings.TrimSuffix(name, filepath.Ext(name))
	item, err := s.repo.Create(r.Context(), library.Item{Title: title, Kind: kind, Status: "planned", Total: total, File: filename, Format: strings.TrimPrefix(ext, ".")})
	if err != nil {
		s.fail(w, err)
		return
	}
	keep = true
	respond(w, 201, item)
}
func (s *Server) local(r *http.Request) (library.Item, string, error) {
	i, err := s.repo.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		return i, "", err
	}
	if i.File == "" || filepath.Base(i.File) != i.File {
		return i, "", library.ErrNotFound
	}
	return i, filepath.Join(s.config.Files, i.File), nil
}
func (s *Server) content(w http.ResponseWriter, r *http.Request) {
	i, path, err := s.local(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	var content reader.Content
	switch i.Format {
	case "epub":
		content, err = reader.EPUB(path)
	case "cbz":
		content, err = reader.Comic(path)
	case "txt":
		var b []byte
		b, err = os.ReadFile(path)
		content.Chapters = []reader.Chapter{{Title: i.Title, Text: string(b)}}
	default:
		problem(w, 400, "Open this item with the file viewer")
		return
	}
	if err != nil {
		problem(w, 422, "The document could not be opened")
		return
	}
	respond(w, 200, content)
}
func (s *Server) file(w http.ResponseWriter, r *http.Request) {
	i, path, err := s.local(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'self'; base-uri 'none'")
	switch i.Format {
	case "pdf":
		w.Header().Set("Content-Type", "application/pdf")
	case "mp4":
		w.Header().Set("Content-Type", "video/mp4")
	case "webm":
		w.Header().Set("Content-Type", "video/webm")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment")
	}
	http.ServeFile(w, r, path)
}
func (s *Server) image(w http.ResponseWriter, r *http.Request) {
	i, path, err := s.local(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	if i.Format != "cbz" {
		problem(w, 400, "Not a comic archive")
		return
	}
	index, _ := strconv.Atoi(r.PathValue("page"))
	b, err := reader.ComicImage(path, index)
	if err != nil {
		problem(w, 404, "Page not found")
		return
	}
	contentType := http.DetectContentType(b)
	if !strings.HasPrefix(contentType, "image/") {
		problem(w, 422, "Invalid image")
		return
	}
	w.Header().Set("Content-Type", contentType)
	if _, err = w.Write(b); err != nil {
		slog.Debug("image response interrupted", "error", err)
	}
}
