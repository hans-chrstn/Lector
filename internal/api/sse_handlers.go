package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/user/lector/internal/models"
	"github.com/user/lector/internal/services"
)

var (
	sseClients = make(map[chan models.Job]bool)
	sseMu      sync.Mutex
)

func init() {
	services.BroadcastJobUpdate = func(job models.Job) {
		sseMu.Lock()
		defer sseMu.Unlock()
		for ch := range sseClients {
			select {
			case ch <- job:
			default:
			}
		}
	}
}

func (h *API) SSEEvents(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	ch := make(chan models.Job, 100)
	sseMu.Lock()
	sseClients[ch] = true
	sseMu.Unlock()

	notify := c.Context().Done()

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer func() {
			sseMu.Lock()
			delete(sseClients, ch)
			sseMu.Unlock()
			close(ch)
		}()

		fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
		w.Flush()

		for {
			select {
			case <-notify:
				return
			case job, ok := <-ch:
				if !ok {
					return
				}
				data, _ := json.Marshal(job)
				fmt.Fprintf(w, "event: job_update\ndata: %s\n\n", string(data))
				if err := w.Flush(); err != nil {
					return
				}
			}
		}
	})

	return nil
}
