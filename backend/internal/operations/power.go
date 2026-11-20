package operations

import (
	"math"

	"sezzle-calculator/backend/internal/domain"
)

// Power raises the first operand (base) to the second (exponent).
//
// A zero base with a negative exponent is reported as undefined rather than
// as +Inf, because it is a division by zero in disguise (0^-n = 1/0^n).
type Power struct{}

// Name implements domain.Operation.
func (Power) Name() string { return "power" }

// Symbol implements domain.Operation.
func (Power) Symbol() string { return "^" }

// Arity implements domain.Operation.
func (Power) Arity() int { return domain.Binary }

// Apply implements domain.Operation.
func (Power) Apply(operands []float64) (float64, error) {
	base, exponent := operands[0], operands[1]
	if base == 0 && exponent < 0 {
		return 0, domain.ErrUndefinedResult
	}
	return math.Pow(base, exponent), nil
}
