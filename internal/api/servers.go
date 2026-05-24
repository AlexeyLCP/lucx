package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/alexeylcp/lucx-core/internal/backend"
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
		id := chi.URLParam(r, "id")
		srv, err := h.Store.GetServer(id)
		if err != nil {
			http.Error(w, `{"error":"not found"}`, 404)
			return
		}
		client, err := h.SSHFactory(srv)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"ssh: %s"}`, err.Error()), 500)
			return
		}
		defer client.Close()

		check, _ := scanner.PreInstallCheck(client)
		if !check.Safe && !check.CanInstall {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, check.Warning), 409)
			return
		}

		be, _ := backend.Get(backend.BackendXray)
		path, err := be.Install(r.Context(), client)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"install: %s"}`, err.Error()), 500)
			return
		}
		if err := be.Start(r.Context(), client); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"start: %s"}`, err.Error()), 500)
			return
		}
		status, _ := be.Status(r.Context(), client)

		h.Store.UpsertServerBackend(&store.ServerBackend{
			ServerID: id, BackendType: "xray", Version: status.Version,
			Status: "running", ConfigPath: path, ConfigManaged: true,
		})
		h.Store.UpdateServerStatus(id, "online")

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"installed","version":"%s","path":"%s"}`, status.Version, path)
	}
}

func (h *Handlers) handleImportServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		srv, err := h.Store.GetServer(id)
		if err != nil {
			http.Error(w, `{"error":"not found"}`, 404)
			return
		}
		client, err := h.SSHFactory(srv)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"ssh: %s"}`, err.Error()), 500)
			return
		}
		defer client.Close()

		cfg, err := scanner.ImportExisting(client)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"import: %s"}`, err.Error()), 500)
			return
		}

		h.Store.UpdateServerStatus(id, "imported")
		h.Store.UpsertServerBackend(&store.ServerBackend{
			ServerID: id, BackendType: "xray", Status: "imported",
			ConfigManaged: false,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg)
	}
}
