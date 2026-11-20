package operations

import "sezzle-calculator/backend/internal/domain"

// Subtract returns the first operand minus the second.
type Subtract struct{}

// Name implements domain.Operation.
func (Subtract) Name() string { return "subtract" }

// Symbol implements domain.Operation.
func (Subtract) Symbol() string { return "-" }

// Arity implements domain.Operation.
func (Subtract) Arity() int { return domain.Binary }

// Apply implements domain.Operation.
func (Subtract) Apply(operands []float64) (float64, error) {
	return operands[0] - operands[1], nil
}
