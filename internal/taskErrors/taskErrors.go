package taskErrors

import (
	"fmt"

	"github.com/google/uuid"
)

type TaskNotFoundError struct {
	ID uuid.UUID
}

func (e TaskNotFoundError) Error() string {
	return fmt.Sprintf("task %s not found", e.ID)
}

var ErrServerBusy = fmt.Errorf("server is busy")

var ErrOverload = fmt.Errorf("task already has 3 files")

var ErrAllFailed = fmt.Errorf("all downloads failed")
