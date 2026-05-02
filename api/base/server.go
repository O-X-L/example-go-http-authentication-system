package base

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example_api/base/db"

	"github.com/go-playground/validator/v10"
)

type Server struct {
	DataStore *db.DataStore
	Mux       *http.ServeMux
	Validator *validator.Validate
}

// NewServer initializes the server, the validator, and registers all routes
func NewServer(store *db.DataStore) *Server {
	srv := &Server{
		DataStore: store,
		Mux:       http.NewServeMux(),
		Validator: validator.New(),
	}
	srv.setupRoutes()
	return srv
}

// Run starts the HTTP server and blocks until a termination signal is received
func (s *Server) Run(addr string) {
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: CORSMiddleware(s.Mux),
	}

	go func() {
		log.Printf("Server starting on %s\n", addr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting gracefully.")
}
