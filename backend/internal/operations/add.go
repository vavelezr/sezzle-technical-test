package operations

import "sezzle-calculator/backend/internal/domain"

// Add returns the sum of two operands.
type Add struct{}

// Name implements domain.Operation.
func (Add) Name() string { return "add" }

// Symbol implements domain.Operation.
func (Add) Symbol() string { return "+" }

// Arity implements domain.Operation.
func (Add) Arity() int { return domain.Binary }

// Apply implements domain.Operation.
func (Add) Apply(operands []float64) (float64, error) {
	return operands[0] + operands[1], nil
}
