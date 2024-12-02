package server

import (
	"context"
	"net/http"
	"time"

	"github.com/alejoacosta74/go-logger"
	"github.com/gin-gonic/gin"
)

type Server struct {
	*http.Server
}

func NewServer(port string) *Server {
	// Configure Gin to run in production mode
	gin.SetMode(gin.ReleaseMode)
	// Create default Gin router with middleware
	router := gin.Default()
	return &Server{
		Server: &http.Server{
			Addr:    port,
			Handler: router,
		},
	}
}

func (s Server) Get(relativePath string, handlers ...gin.HandlerFunc) {
	s.Server.Handler.(*gin.Engine).GET(relativePath, handlers...)
}

func (s Server) Post(relativePath string, handlers ...gin.HandlerFunc) {
	s.Server.Handler.(*gin.Engine).POST(relativePath, handlers...)
}

func (s Server) ListenAndServe() {
	go func() {
		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to run the server", "error", err)
		}
	}()
}

func (s Server) Shutdown(ctx context.Context) {
	ctxWithTimeout, _ := context.WithTimeout(ctx, 5*time.Second)
	if err := s.Server.Shutdown(ctxWithTimeout); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}
	logger.Info("server exiting")
}
