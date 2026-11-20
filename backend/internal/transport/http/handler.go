// Package httptransport exposes the calculator over a JSON REST API using
// only net/http. It owns everything HTTP-specific: routing, DTOs, error
// mapping, middleware and the server lifecycle.
package httptransport

import (
	"context"
	"log/slog"
	"net/http"

	"sezzle-calculator/backend/internal/domain"
)

// Calculator is the use-case port the HTTP layer depends on. It is satisfied
// by *service.CalculatorService and can be replaced with a fake in tests.
type Calculator interface {
	Calculate(ctx context.Context, operation string, operands []float64) (float64, error)
	Operations() []domain.Operation
}

// Handler serves the calculator API endpoints.
type Handler struct {
	calculator   Calculator
	logger       *slog.Logger
	maxBodyBytes int64
}

// NewHandler returns a Handler. Request bodies larger than maxBodyBytes are
// rejected with 413.
func NewHandler(calculator Calculator, logger *slog.Logger, maxBodyBytes int64) *Handler {
	return &Handler{calculator: calculator, logger: logger, maxBodyBytes: maxBodyBytes}
}

// Health reports that the service is up.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.respond(w, r, http.StatusOK, healthResponse{Status: "ok"})
}

// Operations lists the supported operations and their metadata.
func (h *Handler) Operations(w http.ResponseWriter, r *http.Request) {
	h.respond(w, r, http.StatusOK, newOperationsResponse(h.calculator.Operations()))
}

// Calculate decodes a calculation request, runs it and returns the result.
func (h *Handler) Calculate(w http.ResponseWriter, r *http.Request) {
	var req calculateRequest
	if err := decodeJSON(w, r, h.maxBodyBytes, &req); err != nil {
		h.respondError(w, r, err)
		return
	}

	operands, err := req.operands()
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	result, err := h.calculator.Calculate(r.Context(), req.Operation, operands)
	if err != nil {
		h.respondError(w, r, err)
		return
	}

	h.respond(w, r, http.StatusOK, calculateResponse{Operation: req.Operation, Operands: operands, Result: result})
}

// NotFound answers requests for unknown routes.
func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	h.respondError(w, r, newAPIError(http.StatusNotFound, CodeNotFound, "route %s not found", r.URL.Path))
}

// MethodNotAllowed returns a handler that rejects a known route requested
// with the wrong method, advertising the allowed one in the Allow header.
func (h *Handler) MethodNotAllowed(allowed string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", allowed)
		h.respondError(w, r, newAPIError(http.StatusMethodNotAllowed, CodeMethodNotAllowed,
			"method %s is not allowed on %s", r.Method, r.URL.Path))
	}
}

func (h *Handler) respond(w http.ResponseWriter, r *http.Request, status int, body any) {
	if err := writeJSON(w, status, body); err != nil {
		h.logger.WarnContext(r.Context(), "failed to write response",
			slog.String("request_id", RequestIDFromContext(r.Context())), slog.Any("error", err))
	}
}

func (h *Handler) respondError(w http.ResponseWriter, r *http.Request, err error) {
	apiErr := toAPIError(err)
	if apiErr.Status >= http.StatusInternalServerError {
		h.logger.ErrorContext(r.Context(), "request failed",
			slog.String("request_id", RequestIDFromContext(r.Context())), slog.Any("error", err))
	}
	h.respond(w, r, apiErr.Status, newErrorResponse(apiErr))
}
