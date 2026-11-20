package domain_test

import (
	"errors"
	"testing"

	"sezzle-calculator/backend/internal/domain"
)

func TestTypedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		sentinel error
		message  string
	}{
		{
			name:     "unsupported operation",
			err:      &domain.UnsupportedOperationError{Name: "modulo"},
			sentinel: domain.ErrUnsupportedOperation,
			message:  `unsupported operation: "modulo"`,
		},
		{
			name:     "operand count",
			err:      &domain.OperandCountError{Operation: "add", Expected: 2, Got: 1},
			sentinel: domain.ErrInvalidOperandCount,
			message:  `operation "add" expects 2 operand(s), got 1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, tt.sentinel) {
				t.Errorf("errors.Is(%v, %v) = false, want true", tt.err, tt.sentinel)
			}
			if got := tt.err.Error(); got != tt.message {
				t.Errorf("Error() = %q, want %q", got, tt.message)
			}
		})
	}
}
