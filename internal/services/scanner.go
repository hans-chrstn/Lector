package services

import (
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
)

var (
	ScanTotal atomic.Int32
	ScanDone  atomic.Int32
)

type scanJob struct {
	lp   models.LibraryPath
	path string
}

func ScanLibraryPaths() {
	var paths []models.LibraryPath
	if err := db.DB.Find(&paths).Error; err != nil {
		log.Printf("[Scanner] Failed to fetch library paths: %v", err)
		return
	}

	ScanTotal.Store(0)
	ScanDone.Store(0)

	jobs := make(chan scanJob, 1000)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				processFileFromScanner(job.lp, job.path)
				ScanDone.Add(1)
			}
		}()
	}

	for _, lp := range paths {
		log.Printf("[Scanner] Scanning path: %s (Pattern: %s)", lp.Path, lp.Pattern)
		filepath.WalkDir(lp.Path, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			if isSupportedFormat(path) {
				ScanTotal.Add(1)
				jobs <- scanJob{lp: lp, path: path}
			}
			return nil
		})
	}
	
	close(jobs)
	wg.Wait()
	log.Printf("[Scanner] Library scan complete")
}

func isSupportedFormat(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".epub" || ext == ".pdf" || ext == ".cbz" || ext == ".cbr"
}

func processFileFromScanner(lp models.LibraryPath, absPath string) {
	doc, err := ProcessLocalFile(absPath)
	if err != nil {
		log.Printf("[Scanner] Failed to process %s: %v", absPath, err)
		return
	}

	if lp.Pattern == "None/Flat" || lp.Pattern == "" {
		return
	}

	relPath, err := filepath.Rel(lp.Path, absPath)
	if err != nil {
		return
	}

	meta := parseMetadataFromPath(relPath, lp.Pattern)

	needsSave := false
	if doc.Author == "" || doc.Author == "UNKNOWN" {
		if meta["Author"] != "" {
			doc.Author = meta["Author"]
			needsSave = true
		}
	}
	if meta["Group"] != "" {
		var group models.Group
		db.DB.FirstOrCreate(&group, models.Group{Name: meta["Group"]})
		if doc.GroupID != group.ID {
			doc.GroupID = group.ID
			needsSave = true
		}
	}

	if needsSave {
		db.DB.Save(doc)
	}
}

func parseMetadataFromPath(relPath, pattern string) map[string]string {
	res := make(map[string]string)
	relPath = filepath.ToSlash(relPath)
	pattern = strings.TrimPrefix(pattern, "/")
	pathParts := strings.Split(relPath, "/")
	patternParts := strings.Split(pattern, "/")
	for i, pPart := range patternParts {
		if i >= len(pathParts)-1 {
			break
		}
		val := pathParts[i]
		if strings.Contains(pPart, "{Author}") {
			res["Author"] = val
		} else if strings.Contains(pPart, "{Group}") {
			res["Group"] = val
		}
	}
	return res
}

func IsPathAuthorized(absPath string) bool {
	absPath, err := filepath.Abs(filepath.Clean(absPath))
	if err != nil {
		return false
	}
	var paths []models.LibraryPath
	db.DB.Find(&paths)

	checkPrefix := func(path, prefix string) bool {
		if path == prefix {
			return true
		}
		if !strings.HasSuffix(prefix, string(filepath.Separator)) {
			prefix += string(filepath.Separator)
		}
		return strings.HasPrefix(path, prefix)
	}

	for _, lp := range paths {
		cleanLP, _ := filepath.Abs(filepath.Clean(lp.Path))
		if checkPrefix(absPath, cleanLP) {
			return true
		}
	}
	uploadsDir, _ := filepath.Abs("uploads")
	return checkPrefix(absPath, uploadsDir)
}
