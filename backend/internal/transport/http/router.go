package httptransport

import "net/http"

// APIPrefix is the versioned base path of every endpoint.
const APIPrefix = "/api/v1"

// NewRouter maps the API routes to h using Go 1.22+ method patterns.
//
// net/http would answer unknown routes and wrong methods with plain text, so
// both cases are routed explicitly to JSON handlers: each path gets a
// method-less fallback returning 405, and "/" catches everything else as 404.
func NewRouter(h *Handler) http.Handler {
	routes := []struct {
		method  string
		path    string
		handler http.HandlerFunc
	}{
		{http.MethodGet, APIPrefix + "/health", h.Health},
		{http.MethodGet, APIPrefix + "/operations", h.Operations},
		{http.MethodPost, APIPrefix + "/calculate", h.Calculate},
	}

	mux := http.NewServeMux()
	for _, rt := range routes {
		mux.Handle(rt.method+" "+rt.path, rt.handler)
		mux.Handle(rt.path, h.MethodNotAllowed(rt.method))
	}
	mux.HandleFunc("/", h.NotFound)
	return mux
}
