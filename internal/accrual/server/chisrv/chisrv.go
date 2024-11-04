package chisrv

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/zasuchilas/gophermart/internal/accrual/config"
	"github.com/zasuchilas/gophermart/internal/accrual/logger"
	"github.com/zasuchilas/gophermart/internal/accrual/storage"
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

	r.Group(func(r chi.Router) {
		r.Use(middleware.Throttle(3)) // 15
		r.Get("/api/orders/{orderNum}", s.getOrderAccrual)
	})

	return r
}
