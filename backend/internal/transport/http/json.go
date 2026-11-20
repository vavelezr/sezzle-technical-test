package httptransport

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

// writeJSON serializes v as the response body with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// decodeJSON strictly decodes a single JSON object from the request body into
// dst. The body is capped at maxBytes and unknown fields are rejected. Every
// failure is returned as an *apiError with a client-friendly message.
func decodeJSON(w http.ResponseWriter, r *http.Request, maxBytes int64, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return decodeError(err)
	}

	switch err := dec.Decode(&struct{}{}); {
	case errors.Is(err, io.EOF):
		return nil
	case isTooLarge(err):
		return decodeError(err)
	default:
		return newAPIError(http.StatusBadRequest, CodeInvalidRequest, "request body must contain a single JSON object")
	}
}

func decodeError(err error) *apiError {
	var (
		syntaxErr  *json.SyntaxError
		typeErr    *json.UnmarshalTypeError
		maxSizeErr *http.MaxBytesError
	)

	switch {
	case errors.Is(err, io.EOF):
		return newAPIError(http.StatusBadRequest, CodeInvalidRequest, "request body must not be empty")
	case errors.As(err, &maxSizeErr):
		return newAPIError(http.StatusRequestEntityTooLarge, CodePayloadTooLarge,
			"request body must not exceed %d bytes", maxSizeErr.Limit)
	case errors.As(err, &syntaxErr), errors.Is(err, io.ErrUnexpectedEOF):
		return newAPIError(http.StatusBadRequest, CodeInvalidRequest, "request body contains malformed JSON")
	case errors.As(err, &typeErr):
		if typeErr.Field == "" {
			return newAPIError(http.StatusBadRequest, CodeInvalidRequest, "request body must be a JSON object")
		}
		return newAPIError(http.StatusBadRequest, CodeInvalidRequest, "field %q has an invalid value", typeErr.Field)
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		// encoding/json exposes no typed error for unknown fields.
		field := strings.TrimPrefix(err.Error(), "json: unknown field ")
		return newAPIError(http.StatusBadRequest, CodeInvalidRequest, "request body contains unknown field %s", field)
	default:
		return newAPIError(http.StatusBadRequest, CodeInvalidRequest, "request body is invalid")
	}
}

func isTooLarge(err error) bool {
	var maxSizeErr *http.MaxBytesError
	return errors.As(err, &maxSizeErr)
}
