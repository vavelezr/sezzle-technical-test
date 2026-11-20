package operations

import "sezzle-calculator/backend/internal/domain"

// Divide returns the first operand divided by the second.
type Divide struct{}

// Name implements domain.Operation.
func (Divide) Name() string { return "divide" }

// Symbol implements domain.Operation.
func (Divide) Symbol() string { return "÷" }

// Arity implements domain.Operation.
func (Divide) Arity() int { return domain.Binary }

// Apply implements domain.Operation.
func (Divide) Apply(operands []float64) (float64, error) {
	if operands[1] == 0 {
		return 0, domain.ErrDivisionByZero
	}
	return operands[0] / operands[1], nil
}
