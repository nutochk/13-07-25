package service

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/nutochk/13-07-25/internal/config"
	"github.com/nutochk/13-07-25/internal/models"
	"github.com/nutochk/13-07-25/internal/repository"
	"github.com/nutochk/13-07-25/internal/taskErrors"
	"go.uber.org/zap"
)

type Service interface {
	CreateTask() (*models.TaskResponse, error)
	GetTask(id string) (*models.TaskResponse, error)
	AddLinkToTask(id string, url string) (*models.TaskResponse, error)
}

type TaskService struct {
	repo        repository.Repository
	logger      *zap.Logger
	maxFileSize int64 // bytes
	fileTypes   map[string]struct{}
	storage     string
	timeout     time.Duration
	semaphore   chan struct{}
}

func NewTaskService(repo repository.Repository, logger *zap.Logger, cfg config.Config) *TaskService {
	service := &TaskService{repo: repo, logger: logger, maxFileSize: int64(cfg.MaxFileSize) * 1024 * 1024,
		storage: cfg.Storage, timeout: cfg.Timeout}
	service.fileTypes = make(map[string]struct{})
	for _, t := range cfg.FileTypes {
		service.fileTypes[t] = struct{}{}
	}
	service.semaphore = make(chan struct{}, cfg.MaxProcessingTasks)
	return service
}

func (s *TaskService) CreateTask() (*models.TaskResponse, error) {
	id := uuid.New()

	if len(s.semaphore) == cap(s.semaphore) {
		return nil, taskErrors.ErrServerBusy
	}

	task, err := s.repo.CreateTask(id)

	if err != nil {
		s.logger.Error("error create task", zap.Error(err))
		return nil, err
	}
	response := &models.TaskResponse{ID: task.ID, Status: task.Status}
	return response, nil
}

func (s *TaskService) GetTask(id string) (*models.TaskResponse, error) {
	parsedId, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("error parse task id", zap.Error(err))
		return nil, err
	}

	task, err := s.repo.GetTask(parsedId)

	if err != nil {
		s.logger.Error("error get task", zap.Any("id", id), zap.Error(err))
		return nil, err
	}
	response := &models.TaskResponse{ID: task.ID, Status: task.Status, Zip: task.Zip, ErrorMessages: task.ErrorMessages}

	return response, nil
}

func (s *TaskService) AddLinkToTask(id string, url string) (*models.TaskResponse, error) {
	parsedId, err := uuid.Parse(id)
	if err != nil {
		s.logger.Error("error parse task id", zap.Error(err))
		return nil, err
	}

	task, err := s.repo.AddLinkToTask(parsedId, url)

	if err != nil {
		s.logger.Error("error add link to task", zap.Any("id", id), zap.String("url", url), zap.Error(err))
		return nil, err
	}

	if task.Status == models.StatusReady {
		go s.downloadToZip(&task)
	}

	response := &models.TaskResponse{ID: task.ID, Status: task.Status}
	return response, nil
}

func (s *TaskService) downloadToZip(task *models.Task) {
	s.semaphore <- struct{}{}
	defer func() { <-s.semaphore }()

	err := s.repo.UpdateTaskStatus(task.ID, models.StatusProcessing)
	if err != nil {
		s.logger.Error("error update task status", zap.Any("id", task.ID), zap.Error(err))
		return
	}

	zipName := fmt.Sprintf("task_%s.zip", task.ID.String())
	zipPath := filepath.Join(s.storage, zipName)

	var errMessages []string
	if err = os.MkdirAll(s.storage, 0755); err != nil {
		errMessages = append(errMessages, fmt.Sprintf("failed to create storage directory: %v", err))
		err = s.repo.UpdateTaskInfo(task.ID, "", errMessages)
		if err != nil {
			s.logger.Error("error update task info", zap.Any("id", task.ID), zap.Error(err))
		}
		return
	}

	zipFile, err := os.Create(zipPath)
	if err != nil {
		errMessages = append(errMessages, fmt.Sprintf("failed to create zip file: %v", err))
		err = s.repo.UpdateTaskInfo(task.ID, "", errMessages)
		if err != nil {
			s.logger.Error("error update task info", zap.Any("id", task.ID), zap.Error(err))
		}
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, link := range task.Links {
		errMessage := s.downloadFile(link.URL, zipWriter)
		if errMessage != "" {
			errMessages = append(errMessages, errMessage)
		}
	}

	if len(errMessages) == task.FileCount {
		os.Remove(zipPath)
		err = s.repo.UpdateTaskInfo(task.ID, "", errMessages)
		if err != nil {
			s.logger.Error("error update task info", zap.Any("id", task.ID), zap.Error(err))
		}
		return
	}

	err = s.repo.UpdateTaskInfo(task.ID, zipPath, errMessages)
	if err != nil {
		s.logger.Error("error update task info", zap.Any("id", task.ID), zap.Error(err))
	}
}

func (s *TaskService) downloadFile(url string, writer *zip.Writer) string {
	client := http.Client{Timeout: s.timeout}

	resp, err := client.Head(url)
	if err != nil {
		return fmt.Sprintf("cann't check file type in %s", url)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if _, ok := s.fileTypes[contentType]; !ok {
		return fmt.Sprintf("forbidden file type in %s", url)
	}

	if resp.ContentLength > s.maxFileSize {
		return fmt.Sprintf("more than max allowed file size (%d MB) in %s", s.maxFileSize/(1024*1024), url)
	}

	resp, err = client.Get(url)
	if err != nil {
		return fmt.Sprintf("failed to download from %s, error: %v", url, err)
	}
	defer resp.Body.Close()

	fileName := fmt.Sprintf("file_%d%s", time.Now().UnixNano(), getFileExt(contentType))
	zipFile, err := writer.Create(fileName)
	if err != nil {
		return fmt.Sprintf("failed to create file from %s in zip: %v", url, err)

	}

	time.Sleep(1 * time.Minute)

	if _, err = io.Copy(zipFile, resp.Body); err != nil {
		return fmt.Sprintf("failed to save file from %s in zip: %v", url, err)
	}
	return ""
}

func getFileExt(ct string) string {
	switch ct {
	case "image/jpeg":
		return ".jpg"
	case "application/pdf":
		return ".pdf"
	}
	return ".pdf" // недостижимый случай, из-за предварительной проверки типа
}
