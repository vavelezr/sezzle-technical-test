package httptransport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sezzle-calculator/backend/internal/domain"
	"sezzle-calculator/backend/internal/operations"
	"sezzle-calculator/backend/internal/registry"
	"sezzle-calculator/backend/internal/service"
)

const testMaxBodyBytes = 256

var discardLogger = slog.New(slog.DiscardHandler)

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	r, err := registry.New(operations.All()...)
	if err != nil {
		t.Fatalf("registry.New() unexpected error: %v", err)
	}
	return NewRouter(NewHandler(service.NewCalculatorService(r), discardLogger, testMaxBodyBytes))
}

func serve(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeBody[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("invalid JSON body %q: %v", rec.Body.String(), err)
	}
	return v
}

func TestHealth(t *testing.T) {
	rec := serve(newTestRouter(t), http.MethodGet, "/api/v1/health", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := decodeBody[healthResponse](t, rec); got.Status != "ok" {
		t.Errorf("status field = %q, want ok", got.Status)
	}
}

func TestOperationsEndpoint(t *testing.T) {
	rec := serve(newTestRouter(t), http.MethodGet, "/api/v1/operations", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	got := decodeBody[operationsResponse](t, rec)
	if len(got.Operations) != len(operations.All()) {
		t.Fatalf("got %d operations, want %d", len(got.Operations), len(operations.All()))
	}
	want := operationDTO{Name: "sqrt", Symbol: "√", Arity: 1}
	if got.Operations[5] != want {
		t.Errorf("Operations[5] = %+v, want %+v", got.Operations[5], want)
	}
}

func TestCalculateSuccess(t *testing.T) {
	router := newTestRouter(t)

	tests := []struct {
		name string
		body string
		want calculateResponse
	}{
		{"add", `{"operation":"add","operands":[2,3]}`, calculateResponse{"add", []float64{2, 3}, 5}},
		{"subtract", `{"operation":"subtract","operands":[2,3]}`, calculateResponse{"subtract", []float64{2, 3}, -1}},
		{"multiply", `{"operation":"multiply","operands":[-4,2.5]}`, calculateResponse{"multiply", []float64{-4, 2.5}, -10}},
		{"divide", `{"operation":"divide","operands":[10,4]}`, calculateResponse{"divide", []float64{10, 4}, 2.5}},
		{"power", `{"operation":"power","operands":[2,10]}`, calculateResponse{"power", []float64{2, 10}, 1024}},
		{"sqrt", `{"operation":"sqrt","operands":[16]}`, calculateResponse{"sqrt", []float64{16}, 4}},
		{"percentage", `{"operation":"percentage","operands":[20,50]}`, calculateResponse{"percentage", []float64{20, 50}, 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(router, http.MethodPost, "/api/v1/calculate", tt.body)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body)
			}
			got := decodeBody[calculateResponse](t, rec)
			if got.Operation != tt.want.Operation || got.Result != tt.want.Result ||
				len(got.Operands) != len(tt.want.Operands) {
				t.Errorf("response = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestCalculateErrors(t *testing.T) {
	router := newTestRouter(t)
	oversized := `{"operation":"add","operands":[1,2]}` + strings.Repeat(" ", testMaxBodyBytes)

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{"empty body", "", http.StatusBadRequest, CodeInvalidRequest},
		{"malformed JSON", `{"operation":`, http.StatusBadRequest, CodeInvalidRequest},
		{"invalid syntax", `{operation: add}`, http.StatusBadRequest, CodeInvalidRequest},
		{"unknown field", `{"operation":"add","operands":[1,2],"extra":true}`, http.StatusBadRequest, CodeInvalidRequest},
		{"wrong field type", `{"operation":"add","operands":["1",2]}`, http.StatusBadRequest, CodeInvalidRequest},
		{"number out of float64 range", `{"operation":"add","operands":[1e400,2]}`, http.StatusBadRequest, CodeInvalidRequest},
		{"not an object", `[1,2]`, http.StatusBadRequest, CodeInvalidRequest},
		{"multiple JSON values", `{"operation":"add","operands":[1,2]}{}`, http.StatusBadRequest, CodeInvalidRequest},
		{"body too large", oversized, http.StatusRequestEntityTooLarge, CodePayloadTooLarge},
		{"missing operation", `{"operands":[1,2]}`, http.StatusBadRequest, CodeInvalidRequest},
		{"null operand", `{"operation":"add","operands":[null,2]}`, http.StatusBadRequest, CodeInvalidOperand},
		{"unknown operation", `{"operation":"modulo","operands":[1,2]}`, http.StatusBadRequest, CodeUnsupportedOperation},
		{"too few operands", `{"operation":"add","operands":[1]}`, http.StatusBadRequest, CodeInvalidOperandCount},
		{"too many operands", `{"operation":"sqrt","operands":[1,2]}`, http.StatusBadRequest, CodeInvalidOperandCount},
		{"missing operands", `{"operation":"add"}`, http.StatusBadRequest, CodeInvalidOperandCount},
		{"division by zero", `{"operation":"divide","operands":[1,0]}`, http.StatusUnprocessableEntity, CodeDivisionByZero},
		{"negative square root", `{"operation":"sqrt","operands":[-9]}`, http.StatusUnprocessableEntity, CodeNegativeSquareRoot},
		{"overflow", `{"operation":"power","operands":[10,400]}`, http.StatusUnprocessableEntity, CodeResultOverflow},
		{"undefined result", `{"operation":"power","operands":[0,-1]}`, http.StatusUnprocessableEntity, CodeUndefinedResult},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(router, http.MethodPost, "/api/v1/calculate", tt.body)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rec.Code, tt.wantStatus, rec.Body)
			}
			got := decodeBody[errorResponse](t, rec)
			if got.Error.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", got.Error.Code, tt.wantCode)
			}
			if got.Error.Message == "" {
				t.Error("error message is empty")
			}
		})
	}
}

func TestRouting(t *testing.T) {
	router := newTestRouter(t)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantCode   string
		wantAllow  string
	}{
		{"unknown route", http.MethodGet, "/api/v1/unknown", http.StatusNotFound, CodeNotFound, ""},
		{"root", http.MethodGet, "/", http.StatusNotFound, CodeNotFound, ""},
		{"GET calculate", http.MethodGet, "/api/v1/calculate", http.StatusMethodNotAllowed, CodeMethodNotAllowed, "POST"},
		{"DELETE health", http.MethodDelete, "/api/v1/health", http.StatusMethodNotAllowed, CodeMethodNotAllowed, "GET"},
		{"POST operations", http.MethodPost, "/api/v1/operations", http.StatusMethodNotAllowed, CodeMethodNotAllowed, "GET"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(router, tt.method, tt.path, "")

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := decodeBody[errorResponse](t, rec); got.Error.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", got.Error.Code, tt.wantCode)
			}
			if got := rec.Header().Get("Allow"); got != tt.wantAllow {
				t.Errorf("Allow = %q, want %q", got, tt.wantAllow)
			}
		})
	}
}

// stubCalculator lets tests force service-level failures.
type stubCalculator struct{ err error }

func (s stubCalculator) Calculate(context.Context, string, []float64) (float64, error) {
	return 0, s.err
}

func (stubCalculator) Operations() []domain.Operation { return nil }

func TestCalculateUnexpectedErrorsAreLoggedAndHidden(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"internal error", errors.New("boom"), http.StatusInternalServerError, CodeInternal},
		{"timeout", context.DeadlineExceeded, http.StatusServiceUnavailable, CodeRequestTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			router := NewRouter(NewHandler(stubCalculator{err: tt.err}, logger, testMaxBodyBytes))

			rec := serve(router, http.MethodPost, "/api/v1/calculate", `{"operation":"add","operands":[1,2]}`)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			got := decodeBody[errorResponse](t, rec)
			if got.Error.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", got.Error.Code, tt.wantCode)
			}
			if tt.wantCode == CodeInternal && strings.Contains(rec.Body.String(), tt.err.Error()) {
				t.Error("internal error details leaked to the client")
			}
			if !strings.Contains(logs.String(), tt.err.Error()) {
				t.Errorf("error %q not logged; logs: %s", tt.err, logs.String())
			}
		})
	}
}

// failingWriter simulates a client that disconnected mid-response.
type failingWriter struct{ *httptest.ResponseRecorder }

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("connection reset") }

func TestRespondLogsWriteFailures(t *testing.T) {
	var logs bytes.Buffer
	h := NewHandler(stubCalculator{}, slog.New(slog.NewJSONHandler(&logs, nil)), testMaxBodyBytes)

	h.Health(failingWriter{httptest.NewRecorder()}, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))

	if !strings.Contains(logs.String(), "failed to write response") {
		t.Errorf("write failure not logged; logs: %s", logs.String())
	}
}
