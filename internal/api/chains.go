package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/alexeylcp/lucx-core/internal/store"
)

func (h *Handlers) handleListChains() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		chains, _ := h.Store.ListChains()
		json.NewEncoder(w).Encode(chains)
	}
}

func (h *Handlers) handleCreateChain() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c store.Chain
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, `{"error":"invalid json"}`, 400)
			return
		}
		c.ID = uuid.New().String()
		c.Status = "draft"
		for i := range c.Nodes {
			c.Nodes[i].ChainID = c.ID
			c.Nodes[i].Position = i
			if i == 0 {
				c.Nodes[i].Role = "entry"
			} else if i == len(c.Nodes)-1 {
				c.Nodes[i].Role = "exit"
			} else {
				c.Nodes[i].Role = "hop"
			}
		}
		if err := h.Store.CreateChain(&c); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(c)
	}
}

func (h *Handlers) handleGetChain() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		c, err := h.Store.GetChain(id)
		if err != nil {
			http.Error(w, `{"error":"not found"}`, 404)
			return
		}
		json.NewEncoder(w).Encode(c)
	}
}

func (h *Handlers) handleDeleteChain() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		h.Store.DeleteChain(id)
		w.WriteHeader(204)
	}
}

func (h *Handlers) handleValidateChain() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"not implemented"}`, 501)
	}
}

func (h *Handlers) handleApplyChain() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"not implemented"}`, 501)
	}
}

func (h *Handlers) handleRollbackChain() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"not implemented"}`, 501)
	}
}

func (h *Handlers) handleGetChainConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"not implemented"}`, 501)
	}
}
