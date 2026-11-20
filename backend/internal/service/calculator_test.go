package service_test

import (
	"context"
	"errors"
	"math"
	"testing"

	"sezzle-calculator/backend/internal/domain"
	"sezzle-calculator/backend/internal/operations"
	"sezzle-calculator/backend/internal/registry"
	"sezzle-calculator/backend/internal/service"
)

func newService(t *testing.T) *service.CalculatorService {
	t.Helper()
	r, err := registry.New(operations.All()...)
	if err != nil {
		t.Fatalf("registry.New() unexpected error: %v", err)
	}
	return service.NewCalculatorService(r)
}

func TestCalculate(t *testing.T) {
	svc := newService(t)

	tests := []struct {
		name      string
		operation string
		operands  []float64
		want      float64
		wantErr   error
	}{
		{"add", "add", []float64{2, 3}, 5, nil},
		{"subtract", "subtract", []float64{2, 3}, -1, nil},
		{"multiply", "multiply", []float64{2.5, 4}, 10, nil},
		{"divide", "divide", []float64{10, 4}, 2.5, nil},
		{"power", "power", []float64{2, 8}, 256, nil},
		{"sqrt", "sqrt", []float64{81}, 9, nil},
		{"percentage", "percentage", []float64{15, 200}, 30, nil},

		{"unsupported operation", "modulo", []float64{1, 2}, 0, domain.ErrUnsupportedOperation},
		{"too few operands", "add", []float64{1}, 0, domain.ErrInvalidOperandCount},
		{"too many operands", "sqrt", []float64{4, 9}, 0, domain.ErrInvalidOperandCount},
		{"no operands", "multiply", nil, 0, domain.ErrInvalidOperandCount},
		{"NaN operand", "add", []float64{math.NaN(), 1}, 0, domain.ErrInvalidOperand},
		{"infinite operand", "add", []float64{1, math.Inf(-1)}, 0, domain.ErrInvalidOperand},

		{"division by zero", "divide", []float64{1, 0}, 0, domain.ErrDivisionByZero},
		{"negative square root", "sqrt", []float64{-1}, 0, domain.ErrNegativeSquareRoot},
		{"power overflow", "power", []float64{10, 400}, 0, domain.ErrResultOverflow},
		{"multiply overflow", "multiply", []float64{math.MaxFloat64, 10}, 0, domain.ErrResultOverflow},
		{"add overflow", "add", []float64{math.MaxFloat64, math.MaxFloat64}, 0, domain.ErrResultOverflow},
		{"zero to negative power", "power", []float64{0, -1}, 0, domain.ErrUndefinedResult},
		{"negative base fractional exponent", "power", []float64{-8, 0.5}, 0, domain.ErrUndefinedResult},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Calculate(context.Background(), tt.operation, tt.operands)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Calculate(%q, %v) error = %v, want %v", tt.operation, tt.operands, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Calculate(%q, %v) = %v, want %v", tt.operation, tt.operands, got, tt.want)
			}
		})
	}
}

func TestCalculateNormalizesNegativeZero(t *testing.T) {
	svc := newService(t)

	got, err := svc.Calculate(context.Background(), "multiply", []float64{-1, 0})
	if err != nil {
		t.Fatalf("Calculate() unexpected error: %v", err)
	}
	if math.Signbit(got) {
		t.Errorf("Calculate() = %v, want positive zero", got)
	}
}

func TestCalculateOperandCountErrorDetails(t *testing.T) {
	svc := newService(t)

	_, err := svc.Calculate(context.Background(), "divide", []float64{1, 2, 3})

	var countErr *domain.OperandCountError
	if !errors.As(err, &countErr) {
		t.Fatalf("Calculate() error = %v, want *OperandCountError", err)
	}
	want := domain.OperandCountError{Operation: "divide", Expected: 2, Got: 3}
	if *countErr != want {
		t.Errorf("OperandCountError = %+v, want %+v", *countErr, want)
	}
}

func TestCalculateHonorsCancelledContext(t *testing.T) {
	svc := newService(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Calculate(ctx, "add", []float64{1, 2})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Calculate() error = %v, want %v", err, context.Canceled)
	}
}

func TestOperations(t *testing.T) {
	svc := newService(t)

	got := svc.Operations()
	want := operations.All()
	if len(got) != len(want) {
		t.Fatalf("Operations() returned %d items, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Name() != want[i].Name() {
			t.Errorf("Operations()[%d] = %q, want %q", i, got[i].Name(), want[i].Name())
		}
	}
}
