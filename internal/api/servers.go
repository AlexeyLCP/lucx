package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/alexeylcp/lucx-core/internal/scanner"
	"github.com/alexeylcp/lucx-core/internal/store"
)

func (h *Handlers) handleListServers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servers, _ := h.Store.ListServers()
		json.NewEncoder(w).Encode(servers)
	}
}

func (h *Handlers) handleCreateServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var srv store.Server
		if err := json.NewDecoder(r.Body).Decode(&srv); err != nil {
			http.Error(w, `{"error":"invalid json"}`, 400)
			return
		}
		srv.ID = uuid.New().String()
		srv.Status = "unknown"
		srv.Source = "fresh"
		if err := h.Store.CreateServer(&srv); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(srv)
	}
}

func (h *Handlers) handleGetServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		srv, err := h.Store.GetServer(id)
		if err != nil {
			http.Error(w, `{"error":"not found"}`, 404)
			return
		}
		json.NewEncoder(w).Encode(srv)
	}
}

func (h *Handlers) handleDeleteServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		h.Store.DeleteServer(id)
		w.WriteHeader(204)
	}
}

func (h *Handlers) handleScanServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		srv, err := h.Store.GetServer(id)
		if err != nil {
			http.Error(w, `{"error":"not found"}`, 404)
			return
		}
		client, err := h.SSHFactory(srv)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer client.Close()
		check, err := scanner.PreInstallCheck(client)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(check)
	}
}

func (h *Handlers) handleInstallServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: full implementation in later task
		http.Error(w, `{"error":"not implemented"}`, 501)
	}
}

func (h *Handlers) handleImportServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: full implementation in later task
		http.Error(w, `{"error":"not implemented"}`, 501)
	}
}
