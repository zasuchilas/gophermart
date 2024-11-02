package chisrv

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
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
	r.Post("/api/user/register", s.register)
	r.Post("/api/user/login", s.login)

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))

		r.Post("/api/user/orders", s.loadNewOrder)
		r.Get("/api/user/orders", s.getUserOrders)
		r.Get("/api/user/balance", s.getUserBalance)
		r.Post("/api/user/balance/withdraw", s.withdrawFromBalance)
		r.Get("/api/user/withdrawals", s.getWithdrawalList)
	})

	return r
}
