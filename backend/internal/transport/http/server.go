package httptransport

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Server wraps http.Server with sane timeouts and graceful shutdown.
type Server struct {
	httpServer      *http.Server
	logger          *slog.Logger
	shutdownTimeout time.Duration
}

// Option configures a Server (functional options pattern).
type Option func(*Server)

// WithAddr sets the TCP listen address (default ":8080").
func WithAddr(addr string) Option {
	return func(s *Server) { s.httpServer.Addr = addr }
}

// WithReadTimeout sets the maximum duration for reading a request.
func WithReadTimeout(d time.Duration) Option {
	return func(s *Server) { s.httpServer.ReadTimeout = d }
}

// WithWriteTimeout sets the maximum duration before timing out response writes.
func WithWriteTimeout(d time.Duration) Option {
	return func(s *Server) { s.httpServer.WriteTimeout = d }
}

// WithIdleTimeout sets how long keep-alive connections may stay idle.
func WithIdleTimeout(d time.Duration) Option {
	return func(s *Server) { s.httpServer.IdleTimeout = d }
}

// WithShutdownTimeout bounds how long graceful shutdown waits for in-flight requests.
func WithShutdownTimeout(d time.Duration) Option {
	return func(s *Server) { s.shutdownTimeout = d }
}

// WithLogger sets the logger used for lifecycle events and server errors.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Server) {
		s.logger = logger
		s.httpServer.ErrorLog = slog.NewLogLogger(logger.Handler(), slog.LevelError)
	}
}

// NewServer returns a Server serving handler. Defaults are safe for
// production use and can be overridden with options.
func NewServer(handler http.Handler, opts ...Option) *Server {
	s := &Server{
		httpServer: &http.Server{
			Addr:              ":8080",
			Handler:           handler,
			ReadHeaderTimeout: 2 * time.Second,
			ReadTimeout:       5 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		logger:          slog.New(slog.DiscardHandler),
		shutdownTimeout: 10 * time.Second,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Run listens on the configured address and serves until ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.httpServer.Addr, err)
	}
	return s.Serve(ctx, ln)
}

// Serve accepts connections on ln until ctx is cancelled, then shuts down
// gracefully, waiting up to the shutdown timeout for in-flight requests.
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	serveErr := make(chan error, 1)
	go func() { serveErr <- s.httpServer.Serve(ln) }()
	s.logger.InfoContext(ctx, "server listening", slog.String("addr", ln.Addr().String()))

	select {
	case err := <-serveErr:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}

	s.logger.InfoContext(ctx, "shutting down", slog.String("timeout", s.shutdownTimeout.String()))
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.shutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}
	s.logger.InfoContext(ctx, "server stopped")
	return nil
}
