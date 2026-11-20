// Package service holds the calculator use cases.
package service

import (
	"context"
	"math"

	"sezzle-calculator/backend/internal/domain"
)

// OperationRegistry resolves operations by name. It is satisfied by
// *registry.Registry and declared here, where it is consumed, so the service
// depends on an abstraction rather than a concrete type.
type OperationRegistry interface {
	Get(name string) (domain.Operation, error)
	List() []domain.Operation
}

// CalculatorService is a Facade over the operation registry and strategies.
// It exposes a single Calculate entry point and hides operation lookup,
// operand validation and result checking from callers.
type CalculatorService struct {
	registry OperationRegistry
}

// NewCalculatorService returns a CalculatorService backed by the given registry.
func NewCalculatorService(registry OperationRegistry) *CalculatorService {
	return &CalculatorService{registry: registry}
}

// Operations returns the supported operations in display order.
func (s *CalculatorService) Operations() []domain.Operation {
	return s.registry.List()
}

// Calculate resolves the named operation, validates the operands, applies the
// operation and checks that the result is a finite number.
//
// All failures are domain errors (see the domain package), except context
// cancellation, which is returned as ctx.Err().
func (s *CalculatorService) Calculate(ctx context.Context, name string, operands []float64) (float64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	op, err := s.registry.Get(name)
	if err != nil {
		return 0, err
	}
	if err := validateOperands(op, operands); err != nil {
		return 0, err
	}

	result, err := op.Apply(operands)
	if err != nil {
		return 0, err
	}
	return checkResult(result)
}

func validateOperands(op domain.Operation, operands []float64) error {
	if len(operands) != op.Arity() {
		return &domain.OperandCountError{Operation: op.Name(), Expected: op.Arity(), Got: len(operands)}
	}
	for _, v := range operands {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return domain.ErrInvalidOperand
		}
	}
	return nil
}

// checkResult rejects non-finite results and normalizes negative zero to
// zero, so clients never receive "-0".
func checkResult(result float64) (float64, error) {
	switch {
	case math.IsNaN(result):
		return 0, domain.ErrUndefinedResult
	case math.IsInf(result, 0):
		return 0, domain.ErrResultOverflow
	case result == 0:
		return 0, nil
	default:
		return result, nil
	}
}
