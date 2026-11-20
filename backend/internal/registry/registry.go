// Package registry implements the Factory + Registry pattern for operations:
// it resolves an operation name to its Strategy implementation.
package registry

import (
	"errors"
	"fmt"

	"sezzle-calculator/backend/internal/domain"
)

// ErrInvalidRegistration is returned by New when an operation cannot be
// registered (nil, empty name, or duplicated name).
var ErrInvalidRegistration = errors.New("invalid operation registration")

// Registry is an immutable, concurrency-safe lookup of operations by name.
type Registry struct {
	byName  map[string]domain.Operation
	ordered []domain.Operation
}

// New builds a Registry from the given operations, preserving their order
// for listing. It fails fast on nil operations, empty names and duplicates
// so misconfiguration is caught at startup rather than at request time.
func New(ops ...domain.Operation) (*Registry, error) {
	r := &Registry{
		byName:  make(map[string]domain.Operation, len(ops)),
		ordered: make([]domain.Operation, 0, len(ops)),
	}
	for i, op := range ops {
		if err := r.register(op); err != nil {
			return nil, fmt.Errorf("operation at index %d: %w", i, err)
		}
	}
	return r, nil
}

func (r *Registry) register(op domain.Operation) error {
	if op == nil {
		return fmt.Errorf("%w: nil operation", ErrInvalidRegistration)
	}
	name := op.Name()
	if name == "" {
		return fmt.Errorf("%w: empty name", ErrInvalidRegistration)
	}
	if _, exists := r.byName[name]; exists {
		return fmt.Errorf("%w: duplicate name %q", ErrInvalidRegistration, name)
	}
	r.byName[name] = op
	r.ordered = append(r.ordered, op)
	return nil
}

// Get returns the operation registered under name, or an
// *domain.UnsupportedOperationError if there is none.
func (r *Registry) Get(name string) (domain.Operation, error) {
	op, ok := r.byName[name]
	if !ok {
		return nil, &domain.UnsupportedOperationError{Name: name}
	}
	return op, nil
}

// List returns all registered operations in registration order. The returned
// slice is a copy, so callers cannot mutate the registry.
func (r *Registry) List() []domain.Operation {
	return append([]domain.Operation(nil), r.ordered...)
}
