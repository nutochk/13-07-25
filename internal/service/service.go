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
	MaxFileSize int64 // bytes
	FileTypes   map[string]struct{}
	Storage     string
	Timeout     time.Duration
}

func NewTaskService(repo repository.Repository, logger *zap.Logger, cfg config.Config) *TaskService {
	service := &TaskService{repo: repo, logger: logger, MaxFileSize: int64(cfg.MaxFileSize) * 1024 * 1024, Storage: cfg.Storage, Timeout: cfg.Timeout}
	service.FileTypes = make(map[string]struct{})
	for _, t := range cfg.FileTypes {
		service.FileTypes[t] = struct{}{}
	}
	return service
}

func (s *TaskService) CreateTask() (*models.TaskResponse, error) {
	id := uuid.New()
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
	response := &models.TaskResponse{ID: task.ID}
	if task.FileCount == 3 {
		zipPath, errMessages, err := s.downloadToZip(task)
		if err != nil {
			s.logger.Error("error download zip", zap.Error(err))
			return nil, err
		}
		response.Zip = zipPath
		if errMessages != nil {
			response.Status = models.StatusWithError
		} else {
			response.Status = models.StatusCompleted
		}
		response.ErrorMessages = errMessages
	}
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
	response := &models.TaskResponse{ID: task.ID, Status: task.Status}
	return response, nil
}

func (s *TaskService) downloadToZip(task *models.Task) (string, []string, error) {
	zipName := fmt.Sprintf("task_%s.zip", task.ID.String())
	zipPath := filepath.Join(s.Storage, zipName)

	if err := os.MkdirAll(s.Storage, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create storage directory: %v", err)
	}

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create archive file: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	var errMessages []string
	for _, link := range task.Links {
		errMessage := s.downloadFile(link.URL, zipWriter)
		if errMessage != "" {
			errMessages = append(errMessages, errMessage)
		}
	}

	if len(errMessages) == task.FileCount {
		os.Remove(zipPath)
		return "", errMessages, taskErrors.ErrAllFailed
	}
	return zipPath, errMessages, nil
}

func (s *TaskService) downloadFile(url string, writer *zip.Writer) string {
	client := http.Client{Timeout: s.Timeout}

	resp, err := client.Head(url)
	if err != nil {
		return fmt.Sprintf("cann't check file type in %s", url)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if _, ok := s.FileTypes[contentType]; !ok {
		return fmt.Sprintf("forbidden file type in %s", url)
	}

	if resp.ContentLength > s.MaxFileSize {
		return fmt.Sprintf("more than max allowed file size (%d MB) in %s", s.MaxFileSize/(1024*1024), url)
	}

	resp, err = client.Get(url)
	if err != nil {
		return fmt.Sprintf("failed to download from %s, error: %v", url, err)
	}
	defer resp.Body.Close()

	fileName := fmt.Sprintf("file_%d%s", time.Now().UnixNano(), getFileExt(contentType, url))
	zipFile, err := writer.Create(fileName)
	if err != nil {
		return fmt.Sprintf("failed to create file from %s in zip: %v", url, err)

	}

	if _, err = io.Copy(zipFile, resp.Body); err != nil {
		return fmt.Sprintf("failed to save file from %s in zip: %v", url, err)
	}
	return ""
}

func getFileExt(ct string, url string) string {
	switch ct {
	case "image/jpeg":
		return ".jpg"
	case "application/pdf":
		return ".pdf"
	}
	return ".pdf" // недостижимый случай, из-за предварительной проверки типа
}
