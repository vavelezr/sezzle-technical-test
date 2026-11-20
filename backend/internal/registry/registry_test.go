package registry_test

import (
	"errors"
	"testing"

	"sezzle-calculator/backend/internal/domain"
	"sezzle-calculator/backend/internal/operations"
	"sezzle-calculator/backend/internal/registry"
)

// namedOp is a minimal domain.Operation used to exercise registration rules.
type namedOp struct{ name string }

func (o namedOp) Name() string                     { return o.name }
func (namedOp) Symbol() string                     { return "?" }
func (namedOp) Arity() int                         { return domain.Unary }
func (namedOp) Apply(_ []float64) (float64, error) { return 0, nil }

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		ops     []domain.Operation
		wantErr bool
	}{
		{"no operations", nil, false},
		{"default operations", operations.All(), false},
		{"nil operation", []domain.Operation{operations.Add{}, nil}, true},
		{"empty name", []domain.Operation{namedOp{name: ""}}, true},
		{"duplicate name", []domain.Operation{operations.Add{}, namedOp{name: "add"}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := registry.New(tt.ops...)

			if tt.wantErr {
				if !errors.Is(err, registry.ErrInvalidRegistration) {
					t.Fatalf("New() error = %v, want %v", err, registry.ErrInvalidRegistration)
				}
				if r != nil {
					t.Error("New() returned a registry together with an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("New() unexpected error: %v", err)
			}
			if got := len(r.List()); got != len(tt.ops) {
				t.Errorf("List() length = %d, want %d", got, len(tt.ops))
			}
		})
	}
}

func TestGet(t *testing.T) {
	r, err := registry.New(operations.All()...)
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	tests := []struct {
		name    string
		lookup  string
		wantErr bool
	}{
		{"known binary operation", "divide", false},
		{"known unary operation", "sqrt", false},
		{"unknown operation", "modulo", true},
		{"empty name", "", true},
		{"names are case sensitive", "ADD", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, err := r.Get(tt.lookup)

			if tt.wantErr {
				var unsupported *domain.UnsupportedOperationError
				if !errors.As(err, &unsupported) {
					t.Fatalf("Get(%q) error = %v, want *UnsupportedOperationError", tt.lookup, err)
				}
				if unsupported.Name != tt.lookup {
					t.Errorf("error Name = %q, want %q", unsupported.Name, tt.lookup)
				}
				return
			}
			if err != nil {
				t.Fatalf("Get(%q) unexpected error: %v", tt.lookup, err)
			}
			if op.Name() != tt.lookup {
				t.Errorf("Get(%q).Name() = %q", tt.lookup, op.Name())
			}
		})
	}
}

func TestListReturnsCopyInRegistrationOrder(t *testing.T) {
	r, err := registry.New(operations.Sqrt{}, operations.Add{})
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	list := r.List()
	if list[0].Name() != "sqrt" || list[1].Name() != "add" {
		t.Fatalf("List() order = [%s %s], want [sqrt add]", list[0].Name(), list[1].Name())
	}

	list[0] = nil
	if r.List()[0] == nil {
		t.Error("mutating the List() result changed the registry")
	}
}
