package models

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusReady     TaskStatus = "ready for execution"
	StatusCompleted TaskStatus = "completed"
	StatusWithError TaskStatus = "not all files could be downloaded"
)

type Link struct {
	URL string `json:"href"`
}

type Task struct {
	ID        uuid.UUID  `json:"id"`
	Status    TaskStatus `json:"status"`
	Links     []Link     `json:"links"`
	FileCount int        `json:"file_count"`
	CreatedAt time.Time  `json:"created_at"`
}

type TaskResponse struct {
	ID            uuid.UUID  `json:"id"`
	Status        TaskStatus `json:"status"`
	Zip           string     `json:"zip,omitempty"`
	ErrorMessages []string   `json:"error_messages,omitempty"`
}
