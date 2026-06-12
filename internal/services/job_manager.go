package services

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/user/lector/internal/db"
	"github.com/user/lector/internal/models"
	"gorm.io/gorm"
)

type JobHandler func(job *models.Job, updateProgress func(progress int, msg string)) error

type JobManager struct {
	handlers map[string]JobHandler
	mu       sync.Mutex
	worker   *time.Ticker
	stop     chan struct{}
}

var DefaultJobManager *JobManager

func InitJobManager() *JobManager {
	jm := &JobManager{
		handlers: make(map[string]JobHandler),
		worker:   time.NewTicker(5 * time.Second),
		stop:     make(chan struct{}),
	}
	DefaultJobManager = jm

	db.DB.Model(&models.Job{}).Where("status = ?", "running").Update("status", "pending")

	go jm.poll()
	return jm
}

func (jm *JobManager) RegisterHandler(jobType string, handler JobHandler) {
	jm.mu.Lock()
	defer jm.mu.Unlock()
	jm.handlers[jobType] = handler
}

func (jm *JobManager) Enqueue(jobType string, payload interface{}) (*models.Job, error) {
	payloadBytes, _ := json.Marshal(payload)
	job := &models.Job{
		Type:    jobType,
		Status:  "pending",
		Payload: string(payloadBytes),
	}
	if err := db.DB.Create(job).Error; err != nil {
		return nil, err
	}
	go jm.processNext()
	return job, nil
}

func (jm *JobManager) poll() {
	for {
		select {
		case <-jm.worker.C:
			jm.processNext()
		case <-jm.stop:
			return
		}
	}
}

func (jm *JobManager) processNext() {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	var job models.Job
	if err := db.DB.Where("status = ?", "pending").Order("created_at asc").First(&job).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Printf("[JobManager] Error fetching pending job: %v", err)
		}
		return
	}

	handler, exists := jm.handlers[job.Type]
	if !exists {
		job.Status = "failed"
		job.Error = "No handler registered for job type"
		db.DB.Save(&job)
		return
	}

	job.Status = "running"
	db.DB.Save(&job)

	go func(j models.Job) {
		updateFunc := func(progress int, msg string) {
			j.Progress = progress
			j.Message = msg
			db.DB.Save(&j)
			BroadcastJobUpdate(j)
		}

		err := handler(&j, updateFunc)
		if err != nil {
			j.Status = "failed"
			j.Error = err.Error()
		} else {
			j.Status = "completed"
			j.Progress = 100
		}
		db.DB.Save(&j)
		BroadcastJobUpdate(j)

		go jm.processNext()
	}(job)
}

var BroadcastJobUpdate = func(job models.Job) {}
