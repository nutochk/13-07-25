package transport

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nutochk/13-07-25/internal/models"
	"github.com/nutochk/13-07-25/internal/taskErrors"
)

func (server *HttpServer) create(c *gin.Context) {
	task, err := server.service.CreateTask()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (server *HttpServer) getById(c *gin.Context) {
	id := c.Params.ByName("id")
	if id == "" {
		c.String(http.StatusBadRequest, "id not found")
		return
	}
	task, err := server.service.GetTask(id)
	if err != nil {
		if errors.As(err, &taskErrors.TaskNotFoundError{}) {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, task)
}

func (server *HttpServer) addLink(c *gin.Context) {
	id := c.Params.ByName("id")
	if id == "" {
		c.String(http.StatusBadRequest, "id not found")
		return
	}
	var link models.Link
	if err := c.BindJSON(&link); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	task, err := server.service.AddLinkToTask(id, link.URL)
	if err != nil {
		if errors.As(err, &taskErrors.TaskNotFoundError{}) {
			c.String(http.StatusNotFound, err.Error())
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, task)
}
