// Package domain defines the core calculator abstractions: the Operation
// strategy contract and the typed errors shared across layers.
//
// The domain layer is transport-agnostic: it knows nothing about HTTP or JSON.
package domain

// Arity values for the supported kinds of operations.
const (
	Unary  = 1
	Binary = 2
)

// Operation is the Strategy interface implemented by every arithmetic
// operation. Each implementation encapsulates one algorithm behind the same
// contract, so new operations can be added without modifying existing code.
type Operation interface {
	// Name is the stable identifier used by clients (e.g. "add").
	Name() string
	// Symbol is the human-readable symbol (e.g. "+").
	Symbol() string
	// Arity is the exact number of operands Apply expects.
	Arity() int
	// Apply computes the result. Callers must pass exactly Arity() finite
	// operands; validating that is the responsibility of the service layer.
	Apply(operands []float64) (float64, error)
}
