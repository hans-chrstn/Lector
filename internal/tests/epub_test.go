package tests

import (
	"github.com/user/lector/internal/services"
	"strings"
	"testing"
)

func TestEPUBFixer(t *testing.T) {
	files := map[string]bool{
		"OEBPS/images/cover.jpg": true,
		"OEBPS/chapter1.xhtml":   true,
	}
	fixer := &services.EPUBFixer{Files: files}

	t.Run("Fix Image Paths", func(t *testing.T) {
		html := `<html><body><img src="images/cover.jpg"/><img src="ghost.jpg"/></body></html>`
		fixed := fixer.FixChapter(html, "OEBPS/chapter1.xhtml")
		if !contains(fixed, "images/cover.jpg") {
			t.Errorf("Expected cover.jpg to remain")
		}
		if contains(fixed, "ghost.jpg") {
			t.Errorf("Expected ghost.jpg to be removed")
		}
	})

	t.Run("Fix Link Paths", func(t *testing.T) {
		html := `<html><body><a href="chapter1.xhtml">Link</a><a href="ghost.xhtml">Ghost</a></body></html>`
		fixed := fixer.FixChapter(html, "OEBPS/chapter1.xhtml")
		if !contains(fixed, "chapter1.xhtml") {
			t.Errorf("Expected chapter1.xhtml link to remain")
		}
		if contains(fixed, "ghost.xhtml") {
			t.Errorf("Expected ghost.xhtml link to be unwrapped")
		}
	})
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
