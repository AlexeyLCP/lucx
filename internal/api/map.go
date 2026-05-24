package api

import "net/http"

func (h *Handlers) handleGetMap() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"nodes":[],"edges":[]}`))
	}
}
