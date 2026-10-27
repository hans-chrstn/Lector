package reader

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func archive(t *testing.T, files map[string]string) string {
	t.Helper()
	name := filepath.Join(t.TempDir(), "book.zip")
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for name, body := range files {
		entry, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = entry.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	return name
}
func TestEPUBUsesSpineAndDropsExecutableContent(t *testing.T) {
	file := archive(t, map[string]string{"META-INF/container.xml": `<container><rootfiles><rootfile full-path="OEBPS/book.opf"/></rootfiles></container>`, "OEBPS/book.opf": `<package><manifest><item id="one" href="one.xhtml"/><item id="two" href="two.xhtml"/></manifest><spine><itemref idref="two"/><itemref idref="one"/></spine></package>`, "OEBPS/one.xhtml": `<html><body><h1>First</h1><p>A paragraph.</p><script>danger()</script></body></html>`, "OEBPS/two.xhtml": `<html><body><h1>Second</h1><p>Hello world.</p></body></html>`})
	content, err := EPUB(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(content.Chapters) != 2 || content.Chapters[0].Title != "Second" || strings.Contains(content.Chapters[1].Text, "danger") {
		t.Fatalf("invalid EPUB content: %+v", content)
	}
}
func TestComicPagesAndInvalidArchives(t *testing.T) {
	file := archive(t, map[string]string{"02.png": "second", "01.png": "first", "script.html": "ignored"})
	content, err := Comic(file)
	if err != nil || len(content.Chapters) != 2 {
		t.Fatalf("comic: %+v %v", content, err)
	}
	b, err := ComicImage(file, 1)
	if err != nil || string(b) != "first" {
		t.Fatalf("page ordering: %s %v", b, err)
	}
	if _, err = ComicImage(file, 3); err == nil {
		t.Fatal("out of range page accepted")
	}
	if _, err = EPUB(file); err == nil {
		t.Fatal("invalid EPUB accepted")
	}
}
