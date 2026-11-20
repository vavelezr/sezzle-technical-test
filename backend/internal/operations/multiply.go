package operations

import "sezzle-calculator/backend/internal/domain"

// Multiply returns the product of two operands.
type Multiply struct{}

// Name implements domain.Operation.
func (Multiply) Name() string { return "multiply" }

// Symbol implements domain.Operation.
func (Multiply) Symbol() string { return "×" }

// Arity implements domain.Operation.
func (Multiply) Arity() int { return domain.Binary }

// Apply implements domain.Operation.
func (Multiply) Apply(operands []float64) (float64, error) {
	return operands[0] * operands[1], nil
}
