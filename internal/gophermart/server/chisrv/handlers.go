package chisrv

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/theplant/luhn"
	"github.com/zasuchilas/gophermart/internal/gophermart/logger"
	"github.com/zasuchilas/gophermart/internal/gophermart/model"
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
	_, _ = w.Write([]byte("GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service GOPHERMART service "))
}

func (s *ChiServer) register(w http.ResponseWriter, r *http.Request) {

	// decoding request
	var req model.RegisterRequest
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
	var req model.LoginRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
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
		logger.Log.Debug("cannot get login data from db", zap.String("error", err.Error()))
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
	order, err := strconv.Atoi(string(body))
	if err != nil {
		http.Error(w, "the order number must be a number", http.StatusBadRequest)
		return
	}

	// luna validation
	// https://ru.wikipedia.org/wiki/Алгоритм_Луна
	// https://goodcalculators.com/luhn-algorithm-calculator/?Num=18
	if ok := luhn.Valid(order); !ok {
		http.Error(w, "luna validation failed", http.StatusUnprocessableEntity)
		return
	}

	// writing into db

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte((fmt.Sprintf("userID: %d, oredr: %d", userID, order))))
}

func (s *ChiServer) getUserOrders(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte((fmt.Sprintf("userID: %d", userID))))
}

func (s *ChiServer) getUserBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte((fmt.Sprintf("userID: %d", userID))))
}

func (s *ChiServer) withdrawFromBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte((fmt.Sprintf("userID: %d", userID))))
}

func (s *ChiServer) getWithdrawalList(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte((fmt.Sprintf("userID: %d", userID))))
}
