package operations_test

import (
	"errors"
	"math"
	"testing"

	"sezzle-calculator/backend/internal/domain"
	"sezzle-calculator/backend/internal/operations"
)

const tolerance = 1e-9

func TestMetadata(t *testing.T) {
	tests := []struct {
		op     domain.Operation
		name   string
		symbol string
		arity  int
	}{
		{operations.Add{}, "add", "+", domain.Binary},
		{operations.Subtract{}, "subtract", "-", domain.Binary},
		{operations.Multiply{}, "multiply", "×", domain.Binary},
		{operations.Divide{}, "divide", "÷", domain.Binary},
		{operations.Power{}, "power", "^", domain.Binary},
		{operations.Sqrt{}, "sqrt", "√", domain.Unary},
		{operations.Percentage{}, "percentage", "%", domain.Binary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.op.Name(); got != tt.name {
				t.Errorf("Name() = %q, want %q", got, tt.name)
			}
			if got := tt.op.Symbol(); got != tt.symbol {
				t.Errorf("Symbol() = %q, want %q", got, tt.symbol)
			}
			if got := tt.op.Arity(); got != tt.arity {
				t.Errorf("Arity() = %d, want %d", got, tt.arity)
			}
		})
	}
}

func TestApply(t *testing.T) {
	tests := []struct {
		name     string
		op       domain.Operation
		operands []float64
		want     float64
		wantErr  error
	}{
		{"add integers", operations.Add{}, []float64{2, 3}, 5, nil},
		{"add negatives", operations.Add{}, []float64{-2, -3}, -5, nil},
		{"add decimals", operations.Add{}, []float64{0.1, 0.2}, 0.3, nil},

		{"subtract positive result", operations.Subtract{}, []float64{10, 4}, 6, nil},
		{"subtract negative result", operations.Subtract{}, []float64{4, 10}, -6, nil},

		{"multiply integers", operations.Multiply{}, []float64{6, 7}, 42, nil},
		{"multiply by zero", operations.Multiply{}, []float64{123.45, 0}, 0, nil},
		{"multiply signs", operations.Multiply{}, []float64{-3, 4}, -12, nil},

		{"divide exact", operations.Divide{}, []float64{12, 4}, 3, nil},
		{"divide fractional", operations.Divide{}, []float64{10, 4}, 2.5, nil},
		{"divide zero numerator", operations.Divide{}, []float64{0, 5}, 0, nil},
		{"divide by zero", operations.Divide{}, []float64{10, 0}, 0, domain.ErrDivisionByZero},
		{"divide zero by zero", operations.Divide{}, []float64{0, 0}, 0, domain.ErrDivisionByZero},

		{"power positive exponent", operations.Power{}, []float64{2, 10}, 1024, nil},
		{"power negative exponent", operations.Power{}, []float64{2, -2}, 0.25, nil},
		{"power fractional exponent", operations.Power{}, []float64{9, 0.5}, 3, nil},
		{"power zero exponent", operations.Power{}, []float64{5, 0}, 1, nil},
		{"power zero base", operations.Power{}, []float64{0, 3}, 0, nil},
		{"power zero base negative exponent", operations.Power{}, []float64{0, -1}, 0, domain.ErrUndefinedResult},

		{"sqrt perfect square", operations.Sqrt{}, []float64{16}, 4, nil},
		{"sqrt irrational", operations.Sqrt{}, []float64{2}, math.Sqrt2, nil},
		{"sqrt zero", operations.Sqrt{}, []float64{0}, 0, nil},
		{"sqrt negative", operations.Sqrt{}, []float64{-4}, 0, domain.ErrNegativeSquareRoot},

		{"percentage of value", operations.Percentage{}, []float64{20, 50}, 10, nil},
		{"percentage over hundred", operations.Percentage{}, []float64{150, 80}, 120, nil},
		{"percentage negative", operations.Percentage{}, []float64{-10, 200}, -20, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.op.Apply(tt.operands)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Apply(%v) error = %v, want %v", tt.operands, err, tt.wantErr)
			}
			if tt.wantErr == nil && math.Abs(got-tt.want) > tolerance {
				t.Errorf("Apply(%v) = %v, want %v", tt.operands, got, tt.want)
			}
		})
	}
}

// Overflow and NaN detection is the service's job, so strategies return the
// raw IEEE-754 values. These cases document that contract.
func TestApplyReturnsNonFiniteValuesUnchanged(t *testing.T) {
	tests := []struct {
		name     string
		op       domain.Operation
		operands []float64
		check    func(float64) bool
	}{
		{"power overflow", operations.Power{}, []float64{10, 400}, func(v float64) bool { return math.IsInf(v, 1) }},
		{"multiply overflow", operations.Multiply{}, []float64{math.MaxFloat64, 2}, func(v float64) bool { return math.IsInf(v, 1) }},
		{"power negative base fractional exponent", operations.Power{}, []float64{-8, 1.0 / 3}, math.IsNaN},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.op.Apply(tt.operands)
			if err != nil {
				t.Fatalf("Apply(%v) unexpected error: %v", tt.operands, err)
			}
			if !tt.check(got) {
				t.Errorf("Apply(%v) = %v, want non-finite value", tt.operands, got)
			}
		})
	}
}

func TestAll(t *testing.T) {
	want := []string{"add", "subtract", "multiply", "divide", "power", "sqrt", "percentage"}

	all := operations.All()
	if len(all) != len(want) {
		t.Fatalf("All() returned %d operations, want %d", len(all), len(want))
	}
	for i, op := range all {
		if op.Name() != want[i] {
			t.Errorf("All()[%d].Name() = %q, want %q", i, op.Name(), want[i])
		}
	}
}
