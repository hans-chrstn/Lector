package binder

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/user/lector/internal/models"
)

func BindCBZ(doc *models.Document, outPath string) error {
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	imgIdx := 1
	for _, ch := range doc.Chapters {
		var images []string
		if ch.Metadata != "" {
			json.Unmarshal([]byte(ch.Metadata), &images)
		}
		for _, imgURL := range images {
			if imgURL == "" {
				continue
			}

			resp, err := http.Get(imgURL)
			if err != nil {
				continue
			}

			if resp.StatusCode == 200 {
				ext := ".jpg"
				if ct := resp.Header.Get("Content-Type"); ct == "image/png" {
					ext = ".png"
				} else if ct == "image/webp" {
					ext = ".webp"
				}

				filename := fmt.Sprintf("%04d%s", imgIdx, ext)
				writer, err := zipWriter.Create(filename)
				if err == nil {
					io.Copy(writer, resp.Body)
					imgIdx++
				}
			}
			resp.Body.Close()
		}
	}

	return nil
}
