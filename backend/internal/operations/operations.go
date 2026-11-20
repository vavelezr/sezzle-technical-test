// Package operations contains one Strategy implementation per arithmetic
// operation. Each type satisfies domain.Operation.
//
// To add an operation: create a new file with a type implementing
// domain.Operation and append it to All. Nothing else needs to change.
package operations

import "sezzle-calculator/backend/internal/domain"

// All returns every supported operation in display order. It is the single
// place where operations are registered.
func All() []domain.Operation {
	return []domain.Operation{
		Add{},
		Subtract{},
		Multiply{},
		Divide{},
		Power{},
		Sqrt{},
		Percentage{},
	}
}
