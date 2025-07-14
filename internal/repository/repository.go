package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/nutochk/13-07-25/internal/models"
	"github.com/nutochk/13-07-25/internal/taskErrors"
)

type Repository interface {
	CreateTask(id uuid.UUID) (*models.Task, error)
	GetTask(id uuid.UUID) (*models.Task, error)
	AddLinkToTask(id uuid.UUID, url string) (*models.Task, error)
}

type TaskRepository struct {
	tasks map[uuid.UUID]*models.Task
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks: make(map[uuid.UUID]*models.Task),
	}
}

func (tr *TaskRepository) CreateTask(id uuid.UUID) (*models.Task, error) {
	if len(tr.tasks) >= 3 {
		return nil, taskErrors.ErrServerBusy
	}
	task := &models.Task{
		ID:        id,
		Status:    models.StatusPending,
		Links:     make([]models.Link, 0),
		FileCount: 0,
		CreatedAt: time.Now(),
	}

	tr.tasks[id] = task
	return task, nil
}

func (tr *TaskRepository) GetTask(id uuid.UUID) (*models.Task, error) {
	task, exists := tr.tasks[id]
	if !exists {
		return nil, taskErrors.TaskNotFoundError{ID: id}
	}
	return task, nil
}

func (tr *TaskRepository) AddLinkToTask(id uuid.UUID, url string) (*models.Task, error) {
	task, exists := tr.tasks[id]
	if !exists {
		return nil, taskErrors.TaskNotFoundError{ID: id}
	}
	if task.FileCount == 3 {
		return nil, taskErrors.ErrOverload
	}
	task.Links = append(task.Links, models.Link{URL: url})
	task.FileCount++
	if task.FileCount >= 3 {
		task.Status = models.StatusReady
	}
	tr.tasks[id] = task
	return task, nil
}
