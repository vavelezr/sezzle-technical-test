package httptransport

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestNewServerOptions(t *testing.T) {
	s := NewServer(okHandler(),
		WithAddr(":9999"),
		WithReadTimeout(1*time.Second),
		WithWriteTimeout(2*time.Second),
		WithIdleTimeout(3*time.Second),
		WithShutdownTimeout(4*time.Second),
		WithLogger(discardLogger),
	)

	got := []any{s.httpServer.Addr, s.httpServer.ReadTimeout, s.httpServer.WriteTimeout,
		s.httpServer.IdleTimeout, s.shutdownTimeout, s.logger}
	want := []any{":9999", 1 * time.Second, 2 * time.Second, 3 * time.Second, 4 * time.Second, discardLogger}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("setting %d = %v, want %v", i, got[i], want[i])
		}
	}
	if s.httpServer.ErrorLog == nil {
		t.Error("WithLogger did not set the http.Server error log")
	}
}

func TestNewServerDefaults(t *testing.T) {
	s := NewServer(okHandler())

	if s.httpServer.Addr != ":8080" || s.httpServer.ReadHeaderTimeout == 0 || s.shutdownTimeout == 0 {
		t.Errorf("unexpected defaults: addr=%q readHeader=%v shutdown=%v",
			s.httpServer.Addr, s.httpServer.ReadHeaderTimeout, s.shutdownTimeout)
	}
}

func TestServeShutsDownGracefullyOnCancel(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	s := NewServer(okHandler(), WithShutdownTimeout(time.Second))
	ctx, cancel := context.WithCancel(t.Context())

	done := make(chan error, 1)
	go func() { done <- s.Serve(ctx, ln) }()

	resp, err := http.Get("http://" + ln.Addr().String())
	if err != nil {
		t.Fatalf("request to running server failed: %v", err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Serve() = %v, want nil after graceful shutdown", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down in time")
	}
}

func TestServeReturnsListenerErrors(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	_ = ln.Close() // Serve fails immediately on a closed listener.

	if err := NewServer(okHandler()).Serve(t.Context(), ln); err == nil {
		t.Error("Serve() = nil, want error for closed listener")
	}
}

func TestRunFailsOnInvalidAddress(t *testing.T) {
	err := NewServer(okHandler(), WithAddr("invalid-address")).Run(t.Context())
	if err == nil {
		t.Error("Run() = nil, want listen error")
	}
}

func TestRunServesUntilCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if err := NewServer(okHandler(), WithAddr("127.0.0.1:0")).Run(ctx); err != nil {
		t.Errorf("Run() = %v, want nil", err)
	}
}
