package admin

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/kyotomin/ab-router/internal/storage"
)

type Server struct {
	mux   *http.ServeMux
	store storage.Storage
}

func NewServer(store storage.Storage) *Server {
	s := &Server{
		mux:   &http.ServeMux{},
		store: store,
	}

	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) listRules(w http.ResponseWriter, r *http.Request) {
	rules, err := s.store.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJson(w, 200, rules)
}

func (s *Server) createRule(w http.ResponseWriter, r *http.Request) {
	var rule storage.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	rule.ID = uuid.New()

	if err := validatePercent(rule.Percent); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.store.Add(rule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, 200, rule)
}

func (s *Server) updateRule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var rule storage.Rule

	rule.ID = id
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := validatePercent(rule.Percent); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.store.Update(rule); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, storage.NotFoundErr) {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	writeJson(w, 200, rule)
}

func (s *Server) deleteRule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := s.store.Delete(id); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, storage.NotFoundErr) {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func validatePercent(p int) error {
	if p < 0 || p > 100 {
		return errors.New("percent must be between 0 and 100")
	}
	return nil
}

func writeJson(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
