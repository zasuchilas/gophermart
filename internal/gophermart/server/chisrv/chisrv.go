package chisrv

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/zasuchilas/gophermart/internal/gophermart/config"
	"github.com/zasuchilas/gophermart/internal/gophermart/logger"
	"github.com/zasuchilas/gophermart/internal/gophermart/storage"
	"go.uber.org/zap"
	"net/http"
)

type ChiServer struct {
	store storage.Storage
}

func New(s storage.Storage) *ChiServer {
	srv := &ChiServer{
		store: s,
	}

	return srv
}

func (s *ChiServer) Start() {
	logger.Log.Info("Server starts", zap.String("addr", config.RunAddress))
	logger.Log.Fatal(http.ListenAndServe(config.RunAddress, s.router()).Error())
}

func (s *ChiServer) Stop() {
	// TODO: requests cancelling
}

func (s *ChiServer) router() chi.Router {
	r := chi.NewRouter()

	// middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.AllowContentEncoding("deflate", "gzip"))
	r.Use(middleware.Compress(9, "application/json", "text/plain")) // if Accept-Encoding header (gzip, deflate)

	// routes
	r.Get("/", s.home)
	//r.Get("/{shortURL}", s.readURLHandler)
	//r.Get("/ping", s.ping)

	// routes with guard (if there is no valid token returns error 401 Unauthorized)
	//r.Group(func(r chi.Router) {
	//	r.Use(s.secure.GuardMiddleware)
	//	r.Get("/api/user/urls", s.userURLsHandler)
	//	r.Delete("/api/user/urls", s.deleteURLsHandler)
	//})

	// routes with secure cookie (if there is no valid token assigns a new token)
	//r.Group(func(r chi.Router) {
	//	r.Use(s.secure.SecureMiddleware)
	//	r.Post("/", s.writeURLHandler)
	//	r.Post("/api/shorten", s.shortenHandler)
	//	r.Post("/api/shorten/batch", s.shortenBatchHandler)
	//})

	return r
}

func (s *ChiServer) home(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service "))
}
