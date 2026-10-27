package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/lector/internal/extensions"
	"github.com/user/lector/internal/library"
	"github.com/user/lector/internal/storage"
)

func testServer(t *testing.T, config ...Config) http.Handler {
	t.Helper()
	root := t.TempDir()
	repo, err := storage.Open(filepath.Join(root, "library.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	plugins, err := extensions.New(filepath.Join(root, "plugins"), filepath.Join(root, "grants.json"), "unused")
	if err != nil {
		t.Fatal(err)
	}
	c := Config{Files: filepath.Join(root, "files"), MaxUpload: 1 << 20, AllowedHosts: []string{"localhost"}}
	if len(config) > 0 {
		c.User = config[0].User
		c.Password = config[0].Password
	}
	return New(repo, plugins, c)
}

func TestAuthentication(t *testing.T) {
	h := testServer(t, Config{User: "reader", Password: "test-password"})
	for _, tc := range []struct {
		user, password string
		status         int
	}{{"", "", 401}, {"reader", "wrong", 401}, {"reader", "test-password", 200}} {
		r := httptest.NewRequest("GET", "http://localhost/api/items", nil)
		if tc.user != "" {
			r.SetBasicAuth(tc.user, tc.password)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Fatalf("authentication returned %d, want %d", w.Code, tc.status)
		}
	}
}
func call(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "http://localhost"+path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}
func TestLibraryLifecycle(t *testing.T) {
	h := testServer(t)
	w := call(h, "POST", "/api/items", `{"title":"A book","kind":"book","status":"planned","total":5}`)
	if w.Code != 201 {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	var i library.Item
	if err := json.Unmarshal(w.Body.Bytes(), &i); err != nil {
		t.Fatal(err)
	}
	i.Current = 2
	i.Status = "reading"
	body, _ := json.Marshal(i)
	w = call(h, "PUT", "/api/items/"+i.ID, string(body))
	if w.Code != 200 {
		t.Fatalf("update: %d %s", w.Code, w.Body.String())
	}
	if w = call(h, "PUT", "/api/items/"+i.ID, string(body)); w.Code != 409 {
		t.Fatalf("stale update: %d", w.Code)
	}
	if w = call(h, "GET", "/api/history", ""); w.Code != 200 || !strings.Contains(w.Body.String(), `"current":2`) {
		t.Fatalf("history: %s", w.Body.String())
	}
	if w = call(h, "GET", "/api/export", ""); w.Code != 200 || !strings.Contains(w.Header().Get("Content-Disposition"), "attachment") {
		t.Fatal("export failed")
	}
	if w = call(h, "DELETE", "/api/items/"+i.ID, ""); w.Code != 204 {
		t.Fatalf("delete: %d", w.Code)
	}
	if w = call(h, "GET", "/api/items/"+i.ID, ""); w.Code != 404 {
		t.Fatalf("deleted item accessible: %d", w.Code)
	}
}
func TestRequestBoundary(t *testing.T) {
	h := testServer(t)
	for _, body := range []string{`{"title":"","kind":"book","status":"planned"}`, `{"title":"A","kind":"book","status":"planned","unexpected":true}`, `{"title":"A","kind":"book","status":"planned"} {}`} {
		if w := call(h, "POST", "/api/items", body); w.Code != 400 {
			t.Errorf("invalid request accepted: %d %s", w.Code, body)
		}
	}
	for _, test := range []struct{ host, origin, site string }{{"evil.test", "", ""}, {"localhost", "https://evil.test", ""}, {"localhost", "null", ""}, {"localhost", "", "cross-site"}} {
		r := httptest.NewRequest("GET", "http://"+test.host+"/api/items", nil)
		r.Header.Set("Origin", test.origin)
		r.Header.Set("Sec-Fetch-Site", test.site)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 403 {
			t.Errorf("cross origin accepted: %+v %d", test, w.Code)
		}
	}
	w := call(h, "GET", "/api/unknown", "")
	if w.Code != 404 || !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
		t.Fatal("unknown API route did not return JSON 404")
	}
	w = call(h, "GET", "/api/health", "")
	if w.Code != 200 || strings.Contains(w.Header().Get("Content-Security-Policy"), "script-src 'self' 'unsafe-inline'") {
		t.Fatal("health/security headers")
	}
}
func upload(t *testing.T, h http.Handler, name, body string) *httptest.ResponseRecorder {
	t.Helper()
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	p, err := writer.CreateFormFile("file", name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = io.WriteString(p, body); err != nil {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("POST", "http://localhost/api/import", &b)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestLocalImportsAndReader(t *testing.T) {
	h := testServer(t)
	w := upload(t, h, "hello.txt", "Hello, reader.\n\nA second paragraph.")
	if w.Code != 201 {
		t.Fatalf("import: %d %s", w.Code, w.Body.String())
	}
	var i library.Item
	if err := json.Unmarshal(w.Body.Bytes(), &i); err != nil {
		t.Fatal(err)
	}
	w = call(h, "GET", "/api/items/"+i.ID+"/content", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Hello, reader") {
		t.Fatalf("content: %d %s", w.Code, w.Body.String())
	}
	for _, test := range []struct{ name, body string }{{"bad.exe", "x"}, {"empty.txt", ""}, {"invalid.epub", "not an archive"}, {"fake.pdf", "not a pdf"}} {
		if w = upload(t, h, test.name, test.body); w.Code != 400 {
			t.Errorf("invalid import accepted: %s %d", test.name, w.Code)
		}
	}
	if w = upload(t, h, "huge.txt", strings.Repeat("x", (1<<20)+1)); w.Code != 413 {
		t.Fatalf("size limit not enforced: %d", w.Code)
	}
}
