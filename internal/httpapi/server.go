package httpapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/user/lector/internal/extensions"
	"github.com/user/lector/internal/library"
)

type Config struct {
	Files, Public, User, Password string
	MaxUpload                     int64
	AllowedHosts                  []string
}
type Server struct {
	repo    library.Repository
	plugins *extensions.Manager
	config  Config
}

func New(repo library.Repository, plugins *extensions.Manager, c Config) http.Handler {
	if c.MaxUpload <= 0 {
		c.MaxUpload = 100 << 20
	}
	s := &Server{repo: repo, plugins: plugins, config: c}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := repo.Stats(r.Context()); err != nil {
			s.fail(w, err)
			return
		}
		respond(w, 200, map[string]string{"status": "ok", "version": "2"})
	})
	mux.HandleFunc("GET /api/items", s.list)
	mux.HandleFunc("POST /api/items", s.create)
	mux.HandleFunc("GET /api/items/{id}", s.get)
	mux.HandleFunc("PUT /api/items/{id}", s.update)
	mux.HandleFunc("DELETE /api/items/{id}", s.remove)
	mux.HandleFunc("POST /api/import", s.importFile)
	mux.HandleFunc("GET /api/items/{id}/content", s.content)
	mux.HandleFunc("GET /api/items/{id}/file", s.file)
	mux.HandleFunc("GET /api/items/{id}/pages/{page}", s.image)
	mux.HandleFunc("GET /api/stats", func(w http.ResponseWriter, r *http.Request) {
		v, e := repo.Stats(r.Context())
		if e != nil {
			s.fail(w, e)
			return
		}
		respond(w, 200, v)
	})
	mux.HandleFunc("GET /api/history", func(w http.ResponseWriter, r *http.Request) {
		v, e := repo.History(r.Context())
		if e != nil {
			s.fail(w, e)
			return
		}
		respond(w, 200, v)
	})
	mux.HandleFunc("GET /api/export", s.export)
	mux.HandleFunc("GET /api/plugins", s.pluginList)
	mux.HandleFunc("PUT /api/plugins/{id}", s.pluginToggle)
	mux.HandleFunc("POST /api/plugins/{id}/actions/{action}", s.pluginInvoke)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) { problem(w, 404, "Endpoint not found") })
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			problem(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		s.static(w, r)
	})
	return s.secure(mux)
}
func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Debug("response interrupted", "error", err)
	}
}
func problem(w http.ResponseWriter, status int, message string) {
	respond(w, status, map[string]string{"error": message})
}
func (s *Server) fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, library.ErrNotFound), errors.Is(err, extensions.ErrMissing):
		problem(w, 404, "Item not found")
	case errors.Is(err, library.ErrInvalid):
		problem(w, 400, "Check the title, type, status, and progress values")
	case errors.Is(err, library.ErrConflict):
		problem(w, 409, err.Error())
	case errors.Is(err, extensions.ErrDenied):
		problem(w, 403, err.Error())
	case errors.Is(err, extensions.ErrBusy):
		problem(w, 429, err.Error())
	default:
		slog.Error("request failed", "error", err)
		problem(w, 500, "Could not complete the request. Please try again.")
	}
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		problem(w, 415, "Use application/json")
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		problem(w, 400, "Invalid request body")
		return false
	}
	if err := d.Decode(new(any)); err != io.EOF {
		problem(w, 400, "Expected one JSON object")
		return false
	}
	return true
}
func (s *Server) secure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' blob: data:; font-src 'self'; connect-src 'self'; media-src 'self' blob:; frame-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		host := strings.ToLower(r.Host)
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		host = strings.Trim(host, "[]")
		allowed := false
		for _, h := range s.config.AllowedHosts {
			if strings.EqualFold(strings.TrimSpace(h), host) {
				allowed = true
				break
			}
		}
		if !allowed {
			problem(w, 403, "Host is not allowed")
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			u, err := url.Parse(origin)
			if err != nil || u.Host != r.Host || (u.Scheme != "http" && u.Scheme != "https") {
				problem(w, 403, "Cross-origin requests are not allowed")
				return
			}
		}
		if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			problem(w, 403, "Cross-site requests are not allowed")
			return
		}
		if s.config.User != "" {
			user, pass, ok := r.BasicAuth()
			a, b := sha256.Sum256([]byte(user)), sha256.Sum256([]byte(s.config.User))
			c, d := sha256.Sum256([]byte(pass)), sha256.Sum256([]byte(s.config.Password))
			if !ok || subtle.ConstantTimeCompare(a[:], b[:])&subtle.ConstantTimeCompare(c[:], d[:]) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="Lector", charset="UTF-8"`)
				problem(w, 401, "Authentication required")
				return
			}
		}
		defer func() {
			if value := recover(); value != nil {
				slog.Error("request panic", "value", value)
				problem(w, 500, "Request failed")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	v, e := s.repo.List(r.Context(), library.Filter{Query: q.Get("q"), Kind: q.Get("kind"), Status: q.Get("status"), Collection: q.Get("collection"), Sort: q.Get("sort"), Limit: limit, Offset: offset})
	if e != nil {
		s.fail(w, e)
		return
	}
	respond(w, 200, v)
}
func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	v, e := s.repo.Get(r.Context(), r.PathValue("id"))
	if e != nil {
		s.fail(w, e)
		return
	}
	respond(w, 200, v)
}
func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	var i library.Item
	if !decode(w, r, &i) {
		return
	}
	i.File = ""
	i.Format = ""
	v, e := s.repo.Create(r.Context(), i)
	if e != nil {
		s.fail(w, e)
		return
	}
	respond(w, 201, v)
}
func (s *Server) update(w http.ResponseWriter, r *http.Request) {
	var i library.Item
	if !decode(w, r, &i) {
		return
	}
	i.ID = r.PathValue("id")
	v, e := s.repo.Update(r.Context(), i)
	if e != nil {
		s.fail(w, e)
		return
	}
	respond(w, 200, v)
}
func (s *Server) remove(w http.ResponseWriter, r *http.Request) {
	i, e := s.repo.Get(r.Context(), r.PathValue("id"))
	if e != nil {
		s.fail(w, e)
		return
	}
	if e = s.repo.Delete(r.Context(), i.ID); e != nil {
		s.fail(w, e)
		return
	}
	if i.File != "" {
		if e = os.Remove(filepath.Join(s.config.Files, filepath.Base(i.File))); e != nil && !errors.Is(e, os.ErrNotExist) {
			slog.Warn("file cleanup failed", "error", e)
		}
	}
	w.WriteHeader(204)
}
func (s *Server) export(w http.ResponseWriter, r *http.Request) {
	all := []library.Item{}
	for offset := 0; ; offset += 100 {
		page, err := s.repo.List(r.Context(), library.Filter{Limit: 100, Offset: offset, Sort: "added"})
		if err != nil {
			s.fail(w, err)
			return
		}
		all = append(all, page.Items...)
		if offset+len(page.Items) >= page.Total {
			break
		}
	}
	w.Header().Set("Content-Disposition", `attachment; filename="lector-library.json"`)
	respond(w, 200, all)
}
func (s *Server) static(w http.ResponseWriter, r *http.Request) {
	root, err := os.OpenRoot(s.config.Public)
	if err != nil {
		problem(w, 503, "Build the frontend before opening Lector")
		return
	}
	defer root.Close()
	name := strings.TrimPrefix(r.URL.Path, "/")
	if name == "" {
		name = "index.html"
	}
	f, err := root.Open(name)
	if err != nil {
		if filepath.Ext(name) != "" {
			http.NotFound(w, r)
			return
		}
		f, err = root.Open("index.html")
	}
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	if strings.HasPrefix(name, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	http.ServeContent(w, r, f.Name(), info.ModTime(), f)
}
func (s *Server) pluginList(w http.ResponseWriter, _ *http.Request) {
	respond(w, 200, s.plugins.List())
}
func (s *Server) pluginToggle(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Enabled    bool   `json:"enabled"`
		Generation string `json:"generation"`
	}
	if !decode(w, r, &body) {
		return
	}
	if err := s.plugins.SetEnabled(r.PathValue("id"), body.Generation, body.Enabled); err != nil {
		problem(w, 409, err.Error())
		return
	}
	respond(w, 200, s.plugins.List())
}
func (s *Server) pluginInvoke(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Input string `json:"input"`
	}
	if !decode(w, r, &body) {
		return
	}
	v, err := s.plugins.Invoke(r.Context(), r.PathValue("id"), r.PathValue("action"), body.Input)
	if err != nil {
		if errors.Is(err, extensions.ErrDenied) || errors.Is(err, extensions.ErrBusy) || errors.Is(err, extensions.ErrMissing) {
			s.fail(w, err)
		} else {
			problem(w, 422, err.Error())
		}
		return
	}
	respond(w, 200, v)
}
