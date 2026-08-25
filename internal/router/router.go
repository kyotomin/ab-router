package router

import (
	"net/http"

	"github.com/kyotomin/ab-router/internal/storage"
)

type Server struct {
	mux   *http.ServeMux
	store storage.Storage
}

func (s *Server) selectBackend(r *http.Request) (string, error) {

}
