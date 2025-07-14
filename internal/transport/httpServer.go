package transport

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/nutochk/13-07-25/internal/service"
)

type HttpServer struct {
	engine     *gin.Engine
	service    service.Service
	httpServer *http.Server
}

func NewHttpServer(service service.Service) *HttpServer {
	e := gin.Default()
	s := &HttpServer{
		engine:  e,
		service: service,
		httpServer: &http.Server{
			Handler: e,
		},
	}
	s.registerRouters()
	return s
}

func (s *HttpServer) registerRouters() {
	api := s.engine.Group("/api")
	{
		api.POST("/tasks", s.create)
		api.GET("tasks/:id", s.getById)
		api.POST("tasks/:id", s.addLink)
	}
}

func (s *HttpServer) Run(port int) error {
	s.httpServer.Addr = ":" + strconv.Itoa(port)
	return s.httpServer.ListenAndServe()
}

func (s *HttpServer) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
