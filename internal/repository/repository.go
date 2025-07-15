package repository

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nutochk/13-07-25/internal/models"
	"github.com/nutochk/13-07-25/internal/taskErrors"
)

type Repository interface {
	CreateTask(id uuid.UUID) (models.Task, error)
	GetTask(id uuid.UUID) (models.Task, error)
	AddLinkToTask(id uuid.UUID, url string) (models.Task, error)
	UpdateTaskStatus(id uuid.UUID, status models.TaskStatus) error
	UpdateTaskInfo(id uuid.UUID, path string, errs []string) error
}

type TaskRepository struct {
	mu    sync.RWMutex
	tasks map[uuid.UUID]*models.Task
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks: make(map[uuid.UUID]*models.Task),
	}
}

func (tr *TaskRepository) CreateTask(id uuid.UUID) (models.Task, error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	task := &models.Task{
		ID:        id,
		Status:    models.StatusPending,
		Links:     make([]models.Link, 0),
		FileCount: 0,
		CreatedAt: time.Now(),
	}

	tr.tasks[id] = task
	return *task, nil
}

func (tr *TaskRepository) GetTask(id uuid.UUID) (models.Task, error) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	task, exists := tr.tasks[id]
	if !exists {
		return models.Task{}, taskErrors.TaskNotFoundError{ID: id}
	}
	return *task, nil
}

func (tr *TaskRepository) AddLinkToTask(id uuid.UUID, url string) (models.Task, error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	task, exists := tr.tasks[id]
	if !exists {
		return models.Task{}, taskErrors.TaskNotFoundError{ID: id}
	}
	if task.FileCount == 3 {
		return models.Task{}, taskErrors.ErrOverload
	}
	task.Links = append(task.Links, models.Link{URL: url})
	task.FileCount++
	if task.FileCount == 3 {
		task.Status = models.StatusReady
	}
	return *task, nil
}

func (tr *TaskRepository) UpdateTaskStatus(id uuid.UUID, status models.TaskStatus) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	task, exists := tr.tasks[id]
	if !exists {
		return taskErrors.TaskNotFoundError{ID: id}
	}
	task.Status = status
	return nil
}

func (tr *TaskRepository) UpdateTaskInfo(id uuid.UUID, path string, errs []string) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	task, exists := tr.tasks[id]
	if !exists {
		return taskErrors.TaskNotFoundError{ID: id}
	}

	if path != "" {
		task.Zip = path
	}
	task.ErrorMessages = errs
	if len(task.ErrorMessages) == 0 {
		task.Status = models.StatusCompleted
	} else {
		task.Status = models.StatusWithError
	}
	return nil
}
