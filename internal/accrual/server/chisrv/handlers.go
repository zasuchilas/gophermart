package chisrv

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/theplant/luhn"
	"github.com/zasuchilas/gophermart/internal/accrual/logger"
	"github.com/zasuchilas/gophermart/internal/accrual/models"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func (s *ChiServer) home(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ACCRUAL.GOPHERMART"))
}

func (s *ChiServer) getOrderAccrual(w http.ResponseWriter, r *http.Request) {

	orderNum := chi.URLParam(r, "orderNum")
	if orderNum == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// luna validation
	// https://ru.wikipedia.org/wiki/Алгоритм_Луна
	// https://goodcalculators.com/luhn-algorithm-calculator/?Num=18
	number, err := strconv.Atoi(orderNum)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if ok := luhn.Valid(number); !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// reading from db
	orderData, err := s.store.GetOrderData(r.Context(), orderNum)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		logger.Log.Info("cannot get order data from db", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err = enc.Encode(orderData); err != nil {
		logger.Log.Info("error encoding response", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *ChiServer) registerOrder(w http.ResponseWriter, r *http.Request) {

	// decoding request
	var req models.Receipt
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Info("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// getting receipt
	receipt, err := json.Marshal(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// validation
	orderNum := req.Order
	if orderNum == "" {
		http.Error(w, "the order number is required", http.StatusBadRequest)
		return
	}

	// luna validation
	// https://ru.wikipedia.org/wiki/Алгоритм_Луна
	// https://goodcalculators.com/luhn-algorithm-calculator/?Num=18
	number, err := strconv.Atoi(orderNum)
	if err != nil {
		http.Error(w, "the order number must be a number", http.StatusBadRequest)
		return
	}
	if ok := luhn.Valid(number); !ok {
		http.Error(w, "luna validation failed", http.StatusBadRequest)
		return
	}

	// writing into db
	id, err := s.store.RegisterNewOrder(r.Context(), orderNum, string(receipt))
	if id == 0 {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if err != nil {
		logger.Log.Error("failed to write new order into db", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (s *ChiServer) registerGoods(w http.ResponseWriter, r *http.Request) {

	// decoding request
	var req models.GoodsData
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Info("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// validation
	if len(req.Match) < 3 {
		http.Error(w, "the match cannot be less than three characters", http.StatusBadRequest)
		return
	}
	if req.RewardType != "%" && req.RewardType != "pt" {
		http.Error(w, "the reward_type must be '%' or 'pt'", http.StatusBadRequest)
		return
	}
	if req.Reward < 0 {
		http.Error(w, "the reward cannot be less than zero", http.StatusBadRequest)
		return
	}

	// write into db
	id, err := s.store.RegisterNewGoods(r.Context(), req.Match, req.RewardType, req.Reward)
	if id == 0 {
		w.WriteHeader(http.StatusConflict)
		return
	}
	if err != nil {
		logger.Log.Info("failed to write new goods into db", zap.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
