// internal/server/server.go

package server

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"time"

	"github.com/3rr0r-505/Presenz/internal/config"
	"github.com/3rr0r-505/Presenz/internal/middleware"
	"github.com/3rr0r-505/Presenz/internal/routes"
	"github.com/3rr0r-505/Presenz/internal/services"
)

// Server wraps the http.Server so main.go can start/stop it explicitly,
// replacing what Uvicorn handled implicitly in Python.
type Server struct {
	httpServer *http.Server
}

// New builds the full mux — GET / (redirect), GET /attendance/,
// POST /attendance/submit (rate-limited) — wrapped in activity
// middleware across the whole thing, matching main.py's app assembly.
func New(cfg *config.Config, session *services.Session, db *services.DB, ks *services.Killswitch, entry embed.FS) (*Server, error) {
	handler := &routes.AttendanceHandler{
		Cfg:     cfg,
		Session: session,
		DB:      db,
		Entry:   entry,
	}

	limiter, err := middleware.NewRateLimiter(cfg.Security.RateLimit)
	if err != nil {
		return nil, fmt.Errorf("building rate limiter: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /attendance/", handler.ServeEntry)
	mux.HandleFunc("POST /attendance/submit", limiter.Limit(handler.SubmitAttendance))
	mux.Handle("GET /{$}", http.RedirectHandler("/attendance/", http.StatusFound))

	wrapped := middleware.Activity(ks)(mux)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: wrapped,
	}

	return &Server{httpServer: httpServer}, nil
}

// Start begins serving in a goroutine and returns a channel that reports
// ListenAndServe's terminal error (nil on clean Shutdown), so main.go can
// select on it alongside the killswitch.
func (s *Server) Start() <-chan error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
			return
		}
		errCh <- nil
	}()
	return errCh
}

// Shutdown gracefully stops the server, mirroring Uvicorn's should_exit
// + await server_task pattern.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
