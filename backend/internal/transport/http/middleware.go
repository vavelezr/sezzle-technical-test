package httptransport

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Middleware decorates an http.Handler with cross-cutting behavior
// (Decorator pattern). Chained together they form a Chain of Responsibility.
type Middleware func(http.Handler) http.Handler

// Chain wraps h with the given middleware. The first middleware is the
// outermost one, so it sees the request first and the response last.
func Chain(h http.Handler, middleware ...Middleware) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

// RequestIDHeader carries the request correlation ID.
const RequestIDHeader = "X-Request-ID"

const maxRequestIDLength = 64

type requestIDKey struct{}

// RequestID propagates a valid incoming X-Request-ID or generates a new one,
// stores it in the request context and echoes it in the response header.
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(RequestIDHeader)
			if !isValidRequestID(id) {
				id = rand.Text()
			}
			w.Header().Set(RequestIDHeader, id)
			ctx := context.WithValue(r.Context(), requestIDKey{}, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestIDFromContext returns the request ID stored by RequestID, or "".
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// isValidRequestID accepts short IDs made of URL-safe characters, so that
// client-supplied values cannot inject arbitrary content into logs.
func isValidRequestID(id string) bool {
	if id == "" || len(id) > maxRequestIDLength {
		return false
	}
	for _, c := range id {
		isAlnum := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
		if !isAlnum && c != '-' && c != '_' && c != '.' {
			return false
		}
	}
	return true
}

// Logging writes one structured log line per request.
func Logging(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			level := slog.LevelInfo
			if rec.status >= http.StatusInternalServerError {
				level = slog.LevelError
			}
			logger.LogAttrs(r.Context(), level, "http request",
				slog.String("request_id", RequestIDFromContext(r.Context())),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Int("bytes", rec.bytes),
				slog.Float64("duration_ms", float64(time.Since(start).Microseconds())/1000),
			)
		})
	}
}

// statusRecorder captures the status code and body size of a response.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if !r.wroteHeader {
		r.status = status
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	r.wroteHeader = true
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// Recover turns a panic in a downstream handler into a JSON 500 response and
// logs the stack trace, keeping the server alive.
func Recover(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				// http.ErrAbortHandler deliberately aborts the response; let net/http handle it.
				if rec == http.ErrAbortHandler {
					panic(rec)
				}
				logger.ErrorContext(r.Context(), "panic recovered",
					slog.String("request_id", RequestIDFromContext(r.Context())),
					slog.Any("panic", rec),
					slog.String("stack", string(debug.Stack())),
				)
				internal := newAPIError(http.StatusInternalServerError, CodeInternal, "an unexpected error occurred")
				if err := writeJSON(w, internal.Status, newErrorResponse(internal)); err != nil {
					logger.WarnContext(r.Context(), "failed to write panic response", slog.Any("error", err))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// CORS allows cross-origin requests from the given origins. The value "*"
// allows any origin. Preflight requests from allowed origins are answered
// directly with 204.
func CORS(allowedOrigins []string) Middleware {
	allowAll := slices.Contains(allowedOrigins, "*")
	allowedMethods := strings.Join([]string{http.MethodGet, http.MethodPost, http.MethodOptions}, ", ")
	allowedHeaders := strings.Join([]string{"Content-Type", RequestIDHeader}, ", ")
	const preflightMaxAge = 10 * time.Minute

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			header := w.Header()
			if !allowAll {
				header.Add("Vary", "Origin")
			}
			if !allowAll && !slices.Contains(allowedOrigins, origin) {
				next.ServeHTTP(w, r)
				return
			}

			if allowAll {
				header.Set("Access-Control-Allow-Origin", "*")
			} else {
				header.Set("Access-Control-Allow-Origin", origin)
			}

			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				header.Set("Access-Control-Allow-Methods", allowedMethods)
				header.Set("Access-Control-Allow-Headers", allowedHeaders)
				header.Set("Access-Control-Max-Age", strconv.Itoa(int(preflightMaxAge.Seconds())))
				w.WriteHeader(http.StatusNoContent)
				return
			}

			header.Set("Access-Control-Expose-Headers", RequestIDHeader)
			next.ServeHTTP(w, r)
		})
	}
}

// Timeout bounds the request context with a deadline. Handlers observe it
// through r.Context(); the service aborts work once it expires and the error
// mapper reports it as REQUEST_TIMEOUT.
func Timeout(d time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
