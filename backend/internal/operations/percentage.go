package operations

import "sezzle-calculator/backend/internal/domain"

// Percentage computes "a percent of b": (a / 100) * b.
// For example, percentage(20, 50) = 10.
type Percentage struct{}

// Name implements domain.Operation.
func (Percentage) Name() string { return "percentage" }

// Symbol implements domain.Operation.
func (Percentage) Symbol() string { return "%" }

// Arity implements domain.Operation.
func (Percentage) Arity() int { return domain.Binary }

// Apply implements domain.Operation.
func (Percentage) Apply(operands []float64) (float64, error) {
	return operands[0] / 100 * operands[1], nil
}
