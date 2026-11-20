package operations

import (
	"math"

	"sezzle-calculator/backend/internal/domain"
)

// Sqrt returns the principal square root of its single operand.
type Sqrt struct{}

// Name implements domain.Operation.
func (Sqrt) Name() string { return "sqrt" }

// Symbol implements domain.Operation.
func (Sqrt) Symbol() string { return "√" }

// Arity implements domain.Operation.
func (Sqrt) Arity() int { return domain.Unary }

// Apply implements domain.Operation.
func (Sqrt) Apply(operands []float64) (float64, error) {
	if operands[0] < 0 {
		return 0, domain.ErrNegativeSquareRoot
	}
	return math.Sqrt(operands[0]), nil
}
