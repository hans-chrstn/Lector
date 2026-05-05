package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/user/lector/internal/models"
)

func ProcessLocalFile(path string) (*models.Document, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".epub" {
		return processEPUB(path)
	}
	return nil, fmt.Errorf("unsupported format: %s", ext)
}

func stringBetween(str, start, end string) string {
	s := strings.Index(str, start)
	if s == -1 {
		return ""
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return ""
	}
	return str[s : s+e]
}

func EnsureUploadsDir() {
	os.MkdirAll("uploads", 0755)
}
