package httptransport

import (
	"net/http"

	"sezzle-calculator/backend/internal/domain"
)

// The types in this file are the wire format (DTOs) of the API. The adapter
// functions below translate between them and domain types, so the domain
// never depends on JSON tags or HTTP concerns.

type calculateRequest struct {
	Operation string `json:"operation"`
	// Operands are pointers so that JSON null can be told apart from 0.
	Operands []*float64 `json:"operands"`
}

type calculateResponse struct {
	Operation string    `json:"operation"`
	Operands  []float64 `json:"operands"`
	Result    float64   `json:"result"`
}

type operationDTO struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Arity  int    `json:"arity"`
}

type operationsResponse struct {
	Operations []operationDTO `json:"operations"`
}

type healthResponse struct {
	Status string `json:"status"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

// operands validates the request shape and returns the domain operands.
// Semantic checks (arity, finiteness) are left to the service.
func (r calculateRequest) operands() ([]float64, error) {
	if r.Operation == "" {
		return nil, newAPIError(http.StatusBadRequest, CodeInvalidRequest, `field "operation" is required`)
	}
	operands := make([]float64, len(r.Operands))
	for i, v := range r.Operands {
		if v == nil {
			return nil, newAPIError(http.StatusBadRequest, CodeInvalidOperand, "operand at index %d must be a number", i)
		}
		operands[i] = *v
	}
	return operands, nil
}

func newOperationsResponse(ops []domain.Operation) operationsResponse {
	dtos := make([]operationDTO, len(ops))
	for i, op := range ops {
		dtos[i] = operationDTO{Name: op.Name(), Symbol: op.Symbol(), Arity: op.Arity()}
	}
	return operationsResponse{Operations: dtos}
}

func newErrorResponse(e *apiError) errorResponse {
	return errorResponse{Error: errorBody{Code: e.Code, Message: e.Message}}
}
