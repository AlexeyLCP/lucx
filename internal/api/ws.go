package api

import "net/http"

func (h *Handlers) handleWS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"websocket not implemented"}`, 501)
	}
}
