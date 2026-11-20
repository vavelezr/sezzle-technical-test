package httptransport

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestChainOrder(t *testing.T) {
	var calls []string
	trace := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls = append(calls, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	h := Chain(okHandler(), trace("first"), trace("second"), trace("third"))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if got := strings.Join(calls, ","); got != "first,second,third" {
		t.Errorf("call order = %s, want first,second,third", got)
	}
}

func TestRequestID(t *testing.T) {
	tests := []struct {
		name     string
		incoming string
		wantSame bool
	}{
		{"generated when missing", "", false},
		{"propagated when valid", "abc-123_X.y", true},
		{"replaced when it has unsafe characters", "bad id\nINFO forged", false},
		{"replaced when too long", strings.Repeat("a", maxRequestIDLength+1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var seen string
			h := RequestID()(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				seen = RequestIDFromContext(r.Context())
			}))
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.incoming != "" {
				req.Header.Set(RequestIDHeader, tt.incoming)
			}
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			if seen == "" {
				t.Fatal("request ID missing from context")
			}
			if got := rec.Header().Get(RequestIDHeader); got != seen {
				t.Errorf("response header = %q, context = %q", got, seen)
			}
			if (seen == tt.incoming) != tt.wantSame {
				t.Errorf("request ID = %q, incoming = %q, wantSame = %v", seen, tt.incoming, tt.wantSame)
			}
		})
	}
}

func TestRequestIDFromContextWithoutMiddleware(t *testing.T) {
	if got := RequestIDFromContext(t.Context()); got != "" {
		t.Errorf("RequestIDFromContext() = %q, want empty", got)
	}
}

func TestLogging(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantStatus int
		wantBytes  int
		wantLevel  string
	}{
		{
			name:       "implicit 200",
			handler:    func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("hello")) },
			wantStatus: http.StatusOK,
			wantBytes:  5,
			wantLevel:  "INFO",
		},
		{
			name:       "client error",
			handler:    func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusBadRequest) },
			wantStatus: http.StatusBadRequest,
			wantLevel:  "INFO",
		},
		{
			name: "server error keeps first status",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				w.WriteHeader(http.StatusOK)
			},
			wantStatus: http.StatusBadGateway,
			wantLevel:  "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			h := Chain(tt.handler, RequestID(), Logging(logger))

			h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/calculate", nil))

			var entry struct {
				Level     string `json:"level"`
				RequestID string `json:"request_id"`
				Method    string `json:"method"`
				Path      string `json:"path"`
				Status    int    `json:"status"`
				Bytes     int    `json:"bytes"`
			}
			if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
				t.Fatalf("log line is not JSON: %v (%s)", err, logs.String())
			}
			if entry.Status != tt.wantStatus || entry.Bytes != tt.wantBytes || entry.Level != tt.wantLevel {
				t.Errorf("logged status=%d bytes=%d level=%s, want %d %d %s",
					entry.Status, entry.Bytes, entry.Level, tt.wantStatus, tt.wantBytes, tt.wantLevel)
			}
			if entry.Method != http.MethodPost || entry.Path != "/api/v1/calculate" || entry.RequestID == "" {
				t.Errorf("unexpected log entry: %s", logs.String())
			}
		})
	}
}

func TestStatusRecorderUnwrap(t *testing.T) {
	inner := httptest.NewRecorder()
	rec := &statusRecorder{ResponseWriter: inner}

	if rec.Unwrap() != inner {
		t.Error("Unwrap() did not return the wrapped writer")
	}
}

func TestRecover(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	h := Recover(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("something broke")
	}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if got := decodeBody[errorResponse](t, rec); got.Error.Code != CodeInternal {
		t.Errorf("code = %q, want %q", got.Error.Code, CodeInternal)
	}
	if strings.Contains(rec.Body.String(), "something broke") {
		t.Error("panic value leaked to the client")
	}
	if !strings.Contains(logs.String(), "something broke") || !strings.Contains(logs.String(), "stack") {
		t.Errorf("panic not logged with stack; logs: %s", logs.String())
	}
}

func TestRecoverPassesThroughWithoutPanic(t *testing.T) {
	rec := httptest.NewRecorder()
	Recover(discardLogger)(okHandler()).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestRecoverRepanicsOnErrAbortHandler(t *testing.T) {
	h := Recover(discardLogger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	defer func() {
		if rec := recover(); rec != http.ErrAbortHandler {
			t.Errorf("recovered %v, want http.ErrAbortHandler", rec)
		}
	}()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}

func TestRecoverLogsWriteFailure(t *testing.T) {
	var logs bytes.Buffer
	h := Recover(slog.New(slog.NewJSONHandler(&logs, nil)))(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	h.ServeHTTP(failingWriter{httptest.NewRecorder()}, httptest.NewRequest(http.MethodGet, "/", nil))

	if !strings.Contains(logs.String(), "failed to write panic response") {
		t.Errorf("write failure not logged; logs: %s", logs.String())
	}
}

func TestCORS(t *testing.T) {
	const allowed = "http://localhost:5173"

	tests := []struct {
		name          string
		origins       []string
		method        string
		origin        string
		preflight     bool
		wantStatus    int
		wantAllow     string
		wantVary      bool
		wantPreflight bool
	}{
		{"no origin header", []string{allowed}, http.MethodGet, "", false, http.StatusOK, "", false, false},
		{"allowed origin", []string{allowed}, http.MethodPost, allowed, false, http.StatusOK, allowed, true, false},
		{"disallowed origin", []string{allowed}, http.MethodPost, "https://evil.example", false, http.StatusOK, "", true, false},
		{"wildcard", []string{"*"}, http.MethodGet, "https://any.example", false, http.StatusOK, "*", false, false},
		{"allowed preflight", []string{allowed}, http.MethodOptions, allowed, true, http.StatusNoContent, allowed, true, true},
		{"plain OPTIONS is not a preflight", []string{allowed}, http.MethodOptions, allowed, false, http.StatusOK, allowed, true, false},
		{"disallowed preflight", []string{allowed}, http.MethodOptions, "https://evil.example", true, http.StatusOK, "", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/calculate", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.preflight {
				req.Header.Set("Access-Control-Request-Method", http.MethodPost)
			}
			rec := httptest.NewRecorder()

			CORS(tt.origins)(okHandler()).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantAllow {
				t.Errorf("Allow-Origin = %q, want %q", got, tt.wantAllow)
			}
			if got := rec.Header().Get("Vary") == "Origin"; got != tt.wantVary {
				t.Errorf("Vary: Origin present = %v, want %v", got, tt.wantVary)
			}
			if got := rec.Header().Get("Access-Control-Allow-Methods") != ""; got != tt.wantPreflight {
				t.Errorf("Allow-Methods present = %v, want %v", got, tt.wantPreflight)
			}
		})
	}
}

func TestTimeout(t *testing.T) {
	const timeout = time.Second
	var deadline time.Time
	var hasDeadline bool
	h := Timeout(timeout)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		deadline, hasDeadline = r.Context().Deadline()
	}))

	start := time.Now()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	end := time.Now()

	if !hasDeadline {
		t.Fatal("request context has no deadline")
	}
	if deadline.Before(start.Add(timeout)) || deadline.After(end.Add(timeout)) {
		t.Errorf("deadline %v not within [%v, %v]", deadline, start.Add(timeout), end.Add(timeout))
	}
}
