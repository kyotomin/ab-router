package router

import (
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/kyotomin/ab-router/internal/storage"
)

type Server struct {
	mux   *http.ServeMux
	store storage.Storage
}

func NewServer(store storage.Storage) *Server {
	s := &Server{
		mux:   http.NewServeMux(),
		store: store,
	}
	s.mux.HandleFunc("/", s.handle)
	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

const cookieName = "ab_bucket"

func (s *Server) getOrSetBucket(w http.ResponseWriter, r *http.Request) int {
	if c, err := r.Cookie(cookieName); err == nil {
		if b, err := strconv.Atoi(c.Value); err == nil && b >= 0 && b < 100 {
			return b
		}
	}

	bucket := rand.Intn(100)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    strconv.Itoa(bucket),
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 30,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return bucket
}

func (s *Server) selectBackend(bucket int) (string, error) {
	rules, err := s.store.GetAll()
	if err != nil {
		return "", err
	}

	cursor := 0
	for _, rule := range rules {
		cursor += rule.Percent
		if bucket < cursor {
			return rule.Backend, nil
		}
	}

	return "", storage.NotFoundErr
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	bucket := s.getOrSetBucket(w, r)

	backend, err := s.selectBackend(bucket)
	if err != nil {
		http.Error(w, "no backend available", http.StatusServiceUnavailable)
		return
	}

	target, err := url.Parse(backend)
	if err != nil {
		http.Error(w, "bad backend config", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ServeHTTP(w, r)
}
