package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/services"
	"gorm.io/gorm/clause"
)

func RegisterJobs(h *API) {
	jm := services.DefaultJobManager

	jm.RegisterHandler("batch_refresh", func(job *models.Job, update func(int, string)) error {
		var payload struct {
			IDs []int `json:"ids"`
		}
		if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
			return err
		}

		total := len(payload.IDs)
		if total == 0 {
			return nil
		}

		for i, id := range payload.IDs {
			pct := int((float64(i) / float64(total)) * 100)
			update(pct, fmt.Sprintf("Refreshing %d of %d", i+1, total))

			doc, err := h.DocumentService.GetByID(uint(id))
			if err != nil {
				continue
			}

			if doc.Source != "local" {
				if s, ok := h.Engine.Plugins[doc.Source]; ok {
					if fetched, err := s.GetDocument(doc.URL); err == nil && fetched.Title != "" {
						doc.CoverURL = fetched.CoverURL
						doc.Author = fetched.Author
						doc.Synopsis = fetched.Synopsis
						db.DB.Save(doc)

						for i := range fetched.Chapters {
							fetched.Chapters[i].DocumentID = doc.ID
							fetched.Chapters[i].ID = 0
							fetched.Chapters[i].Order = i + 1
						}
						db.DB.Clauses(clause.OnConflict{
							Columns:   []clause.Column{{Name: "document_id"}, {Name: "url"}},
							DoUpdates: clause.AssignmentColumns([]string{"title", "order_val"}),
						}).CreateInBatches(fetched.Chapters, 100)
					}
				}
			} else {
				services.ProcessLocalFile(doc.LocalPath)
			}

			time.Sleep(500 * time.Millisecond)
		}

		update(100, "Refresh complete")
		return nil
	})

	jm.RegisterHandler("scan_library", func(job *models.Job, update func(int, string)) error {
		services.ScanLibraryPaths(update)
		return nil
	})
}
