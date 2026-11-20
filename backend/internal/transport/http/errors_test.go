package httptransport

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"sezzle-calculator/backend/internal/domain"
)

func TestToAPIError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantStatus  int
		wantCode    string
		wantMessage string
	}{
		{
			name:        "api error passes through",
			err:         newAPIError(http.StatusNotFound, CodeNotFound, "missing"),
			wantStatus:  http.StatusNotFound,
			wantCode:    CodeNotFound,
			wantMessage: "missing",
		},
		{
			name:        "unsupported operation",
			err:         &domain.UnsupportedOperationError{Name: "mod"},
			wantStatus:  http.StatusBadRequest,
			wantCode:    CodeUnsupportedOperation,
			wantMessage: `unsupported operation: "mod"`,
		},
		{
			name:        "invalid operand count",
			err:         &domain.OperandCountError{Operation: "sqrt", Expected: 1, Got: 2},
			wantStatus:  http.StatusBadRequest,
			wantCode:    CodeInvalidOperandCount,
			wantMessage: `operation "sqrt" expects 1 operand(s), got 2`,
		},
		{
			name:        "invalid operand",
			err:         domain.ErrInvalidOperand,
			wantStatus:  http.StatusBadRequest,
			wantCode:    CodeInvalidOperand,
			wantMessage: domain.ErrInvalidOperand.Error(),
		},
		{
			name:        "division by zero",
			err:         domain.ErrDivisionByZero,
			wantStatus:  http.StatusUnprocessableEntity,
			wantCode:    CodeDivisionByZero,
			wantMessage: "division by zero is not allowed",
		},
		{
			name:        "negative square root",
			err:         domain.ErrNegativeSquareRoot,
			wantStatus:  http.StatusUnprocessableEntity,
			wantCode:    CodeNegativeSquareRoot,
			wantMessage: domain.ErrNegativeSquareRoot.Error(),
		},
		{
			name:        "result overflow",
			err:         domain.ErrResultOverflow,
			wantStatus:  http.StatusUnprocessableEntity,
			wantCode:    CodeResultOverflow,
			wantMessage: domain.ErrResultOverflow.Error(),
		},
		{
			name:        "wrapped undefined result",
			err:         fmt.Errorf("context: %w", domain.ErrUndefinedResult),
			wantStatus:  http.StatusUnprocessableEntity,
			wantCode:    CodeUndefinedResult,
			wantMessage: "context: result is undefined",
		},
		{
			name:        "deadline exceeded",
			err:         context.DeadlineExceeded,
			wantStatus:  http.StatusServiceUnavailable,
			wantCode:    CodeRequestTimeout,
			wantMessage: "the request timed out",
		},
		{
			name:        "canceled",
			err:         context.Canceled,
			wantStatus:  http.StatusServiceUnavailable,
			wantCode:    CodeRequestTimeout,
			wantMessage: "the request timed out",
		},
		{
			name:        "unknown error is hidden",
			err:         errors.New("database password is hunter2"),
			wantStatus:  http.StatusInternalServerError,
			wantCode:    CodeInternal,
			wantMessage: "an unexpected error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toAPIError(tt.err)

			if got.Status != tt.wantStatus || got.Code != tt.wantCode || got.Message != tt.wantMessage {
				t.Errorf("toAPIError(%v) = {%d %s %q}, want {%d %s %q}",
					tt.err, got.Status, got.Code, got.Message, tt.wantStatus, tt.wantCode, tt.wantMessage)
			}
		})
	}
}

func TestAPIErrorImplementsError(t *testing.T) {
	err := newAPIError(http.StatusBadRequest, CodeInvalidRequest, "field %q is bad", "x")
	if got := err.Error(); got != `field "x" is bad` {
		t.Errorf("Error() = %q", got)
	}
}
