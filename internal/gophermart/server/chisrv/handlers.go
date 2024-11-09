package chisrv

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/Rhymond/go-money"
	"github.com/theplant/luhn"
	"github.com/zasuchilas/gophermart/internal/gophermart/logger"
	"github.com/zasuchilas/gophermart/internal/gophermart/models"
	"github.com/zasuchilas/gophermart/internal/gophermart/storage"
	"github.com/zasuchilas/gophermart/pkg/passhash"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"time"
)

func (s *ChiServer) home(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("GOPHERMART"))
}

func (s *ChiServer) register(w http.ResponseWriter, r *http.Request) {

	// decoding request
	var req models.RegisterRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// validation
	if len(req.Login) < 3 {
		http.Error(w, "login is shorter than 3", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 6 {
		http.Error(w, "password is shorter than 6", http.StatusBadRequest)
		return
	}

	// make password hash
	pass, err := passhash.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "failed to create a password hash", http.StatusInternalServerError)
		return
	}

	// write into db
	userID, err := s.store.Register(r.Context(), req.Login, pass)
	if userID == 0 {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if err != nil {
		logger.Log.Error("failed to write new user into db", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// authorize user
	setJWTCookie(w, userID)

	w.WriteHeader(http.StatusOK)
}

func (s *ChiServer) login(w http.ResponseWriter, r *http.Request) {

	// decoding request
	var req models.LoginRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Info("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// validation
	if len(req.Login) < 3 || len(req.Password) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get data from db
	loginData, err := s.store.GetLoginData(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		logger.Log.Info("cannot get login data from db", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// check password hash
	ok := passhash.CheckPasswordHash(req.Password, loginData.PasswordHash)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// authorize user
	setJWTCookie(w, loginData.UserID)

	w.WriteHeader(http.StatusOK)
}

func setJWTCookie(w http.ResponseWriter, userID int64) {
	token := makeToken(userID)
	http.SetCookie(w, &http.Cookie{
		//HttpOnly: true,
		Expires: time.Now().Add(100 * 24 * time.Hour),
		//SameSite: http.SameSiteLaxMode,
		// Secure: true,
		Name:  "jwt", // Must be named "jwt" or else the token cannot be searched for by jwtauth.Verifier.
		Value: token,
	})
}

func (s *ChiServer) loadNewOrder(w http.ResponseWriter, r *http.Request) {

	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// decoding request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	orderNum := string(body)

	// luna validation
	// https://ru.wikipedia.org/wiki/Алгоритм_Луна
	// https://goodcalculators.com/luhn-algorithm-calculator/?Num=18
	number, err := strconv.Atoi(string(orderNum))
	if err != nil {
		http.Error(w, "the order number must be a number string", http.StatusBadRequest)
		return
	}
	if ok := luhn.Valid(number); !ok {
		http.Error(w, "luna validation failed", http.StatusUnprocessableEntity)
		return
	}

	// writing into db
	err = s.store.RegisterOrder(r.Context(), userID, orderNum)
	if err != nil {
		if errors.Is(err, storage.ErrNumberDone) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(err, storage.ErrNumberAdded) {
			w.WriteHeader(http.StatusConflict)
		}
		logger.Log.Info("writing into db", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *ChiServer) getUserOrders(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug("getUserOrders starting")

	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	logger.Log.Debug("userID received", zap.Int64("userID", userID))

	// reading from db
	orders, err := s.store.GetUserOrders(r.Context(), userID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		logger.Log.Info("reading from db", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Log.Debug("orders received from pg", zap.Any("orders", orders))

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err = enc.Encode(orders); err != nil {
		logger.Log.Info("error encoding response", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Log.Debug("handler finished the job")
}

func (s *ChiServer) getUserBalance(w http.ResponseWriter, r *http.Request) {

	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// reading from db
	balance, err := s.store.GetUserBalance(r.Context(), userID)
	if err != nil {
		logger.Log.Info("reading from db", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err = enc.Encode(balance); err != nil {
		logger.Log.Info("error encoding response", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *ChiServer) withdrawFromBalance(w http.ResponseWriter, r *http.Request) {

	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// decoding request
	var req models.WithdrawRequest
	dec := json.NewDecoder(r.Body)
	if err = dec.Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// luna validation
	// https://ru.wikipedia.org/wiki/Алгоритм_Луна
	// https://goodcalculators.com/luhn-algorithm-calculator/?Num=18
	orderNum := req.Order
	number, err := strconv.Atoi(orderNum)
	if err != nil {
		http.Error(w, "the order number must be a number string", http.StatusUnprocessableEntity)
		return
	}
	if ok := luhn.Valid(number); !ok {
		http.Error(w, "luna validation failed", http.StatusUnprocessableEntity)
		return
	}

	// money validation and transform
	sum := money.NewFromFloat(req.Sum, money.RUB) //.Amount()
	if sum.IsZero() || sum.IsNegative() {
		http.Error(w, "the sum must be a positive number", http.StatusBadRequest)
		return
	}

	// write into db
	err = s.store.WithdrawTransaction(r.Context(), userID, orderNum, sum)
	if err != nil {
		if errors.Is(err, storage.ErrNotEnoughFunds) {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		logger.Log.Info("writing into db", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *ChiServer) getWithdrawalList(w http.ResponseWriter, r *http.Request) {

	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// reading from db
	withdrawals, err := s.store.GetUserWithdrawals(r.Context(), userID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		logger.Log.Info("reading from db", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err = enc.Encode(withdrawals); err != nil {
		logger.Log.Info("error encoding response", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
