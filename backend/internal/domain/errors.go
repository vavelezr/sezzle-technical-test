package domain

import (
	"errors"
	"fmt"
)

// Sentinel errors describing every failure the calculator can report.
// Callers should match them with errors.Is.
var (
	// ErrUnsupportedOperation is returned when no operation matches a name.
	ErrUnsupportedOperation = errors.New("unsupported operation")
	// ErrInvalidOperandCount is returned when the number of operands does
	// not match the operation's arity.
	ErrInvalidOperandCount = errors.New("invalid number of operands")
	// ErrInvalidOperand is returned when an operand is NaN or infinite.
	ErrInvalidOperand = errors.New("operands must be finite numbers")
	// ErrDivisionByZero is returned when dividing by zero.
	ErrDivisionByZero = errors.New("division by zero is not allowed")
	// ErrNegativeSquareRoot is returned for the square root of a negative number.
	ErrNegativeSquareRoot = errors.New("square root of a negative number is not allowed")
	// ErrResultOverflow is returned when the result exceeds the float64 range.
	ErrResultOverflow = errors.New("result is too large to represent")
	// ErrUndefinedResult is returned when the result is mathematically undefined.
	ErrUndefinedResult = errors.New("result is undefined")
)

// UnsupportedOperationError carries the operation name that could not be
// resolved. It matches ErrUnsupportedOperation via errors.Is.
type UnsupportedOperationError struct {
	Name string
}

// Error implements the error interface.
func (e *UnsupportedOperationError) Error() string {
	return fmt.Sprintf("%s: %q", ErrUnsupportedOperation, e.Name)
}

// Unwrap exposes the sentinel so errors.Is(err, ErrUnsupportedOperation) works.
func (e *UnsupportedOperationError) Unwrap() error {
	return ErrUnsupportedOperation
}

// OperandCountError describes an arity mismatch. It matches
// ErrInvalidOperandCount via errors.Is.
type OperandCountError struct {
	Operation string
	Expected  int
	Got       int
}

// Error implements the error interface.
func (e *OperandCountError) Error() string {
	return fmt.Sprintf("operation %q expects %d operand(s), got %d", e.Operation, e.Expected, e.Got)
}

// Unwrap exposes the sentinel so errors.Is(err, ErrInvalidOperandCount) works.
func (e *OperandCountError) Unwrap() error {
	return ErrInvalidOperandCount
}
