package httptransport

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"sezzle-calculator/backend/internal/domain"
)

// Error codes returned in the "code" field of error responses. They are part
// of the public API contract and must remain stable.
const (
	CodeInvalidRequest       = "INVALID_REQUEST"
	CodePayloadTooLarge      = "PAYLOAD_TOO_LARGE"
	CodeUnsupportedOperation = "UNSUPPORTED_OPERATION"
	CodeInvalidOperandCount  = "INVALID_OPERAND_COUNT"
	CodeInvalidOperand       = "INVALID_OPERAND"
	CodeDivisionByZero       = "DIVISION_BY_ZERO"
	CodeNegativeSquareRoot   = "NEGATIVE_SQUARE_ROOT"
	CodeResultOverflow       = "RESULT_OVERFLOW"
	CodeUndefinedResult      = "UNDEFINED_RESULT"
	CodeNotFound             = "NOT_FOUND"
	CodeMethodNotAllowed     = "METHOD_NOT_ALLOWED"
	CodeRequestTimeout       = "REQUEST_TIMEOUT"
	CodeInternal             = "INTERNAL_ERROR"
)

// apiError is an error already expressed in HTTP terms. Transport-level
// failures (bad JSON, unknown routes...) are created directly as apiErrors.
type apiError struct {
	Status  int
	Code    string
	Message string
}

func (e *apiError) Error() string { return e.Message }

func newAPIError(status int, code, format string, args ...any) *apiError {
	return &apiError{Status: status, Code: code, Message: fmt.Sprintf(format, args...)}
}

// domainErrorMappings associates each domain sentinel with its HTTP
// representation. Input problems are 400; well-formed requests whose maths
// cannot be carried out are 422.
var domainErrorMappings = []struct {
	target error
	status int
	code   string
}{
	{domain.ErrUnsupportedOperation, http.StatusBadRequest, CodeUnsupportedOperation},
	{domain.ErrInvalidOperandCount, http.StatusBadRequest, CodeInvalidOperandCount},
	{domain.ErrInvalidOperand, http.StatusBadRequest, CodeInvalidOperand},
	{domain.ErrDivisionByZero, http.StatusUnprocessableEntity, CodeDivisionByZero},
	{domain.ErrNegativeSquareRoot, http.StatusUnprocessableEntity, CodeNegativeSquareRoot},
	{domain.ErrResultOverflow, http.StatusUnprocessableEntity, CodeResultOverflow},
	{domain.ErrUndefinedResult, http.StatusUnprocessableEntity, CodeUndefinedResult},
}

// toAPIError is the single place where errors are translated into HTTP
// status codes and API error codes. Unknown errors become a generic 500 so
// internal details never leak to clients.
func toAPIError(err error) *apiError {
	var apiErr *apiError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	for _, m := range domainErrorMappings {
		if errors.Is(err, m.target) {
			return &apiError{Status: m.status, Code: m.code, Message: err.Error()}
		}
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return newAPIError(http.StatusServiceUnavailable, CodeRequestTimeout, "the request timed out")
	}
	return newAPIError(http.StatusInternalServerError, CodeInternal, "an unexpected error occurred")
}
