package chisrv

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/zasuchilas/gophermart/internal/accrual/config"
	"github.com/zasuchilas/gophermart/internal/accrual/logger"
	"github.com/zasuchilas/gophermart/internal/accrual/storage"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

type ChiServer struct {
	store     storage.Storage
	waitGroup *sync.WaitGroup
}

func New(s storage.Storage, wg *sync.WaitGroup) *ChiServer {
	srv := &ChiServer{
		store:     s,
		waitGroup: wg,
	}
	return srv
}

func (s *ChiServer) Start() {
	logger.Log.Info("Server starts", zap.String("addr", config.RunAddress))
	logger.Log.Fatal(http.ListenAndServe(config.RunAddress, s.router()).Error())
}

func (s *ChiServer) Stop() {
	// TODO: requests cancelling
	s.waitGroup.Done()
}

func (s *ChiServer) router() chi.Router {
	r := chi.NewRouter()

	// middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.AllowContentEncoding("deflate", "gzip"))
	r.Use(middleware.Compress(9, "application/json", "text/plain")) // if Accept-Encoding header (gzip, deflate)

	// routes
	r.Get("/", s.home)
	r.Post("/api/orders", s.registerOrder)
	r.Post("/api/goods", s.registerGoods)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Throttle(15))
		r.Get("/api/orders/{orderNum}", s.getOrderAccrual)
	})

	return r
}
