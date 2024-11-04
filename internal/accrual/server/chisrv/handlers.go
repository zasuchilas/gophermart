package chisrv

import (
	"github.com/go-chi/chi/v5"
	"github.com/theplant/luhn"
	"net/http"
	"strconv"
)

func (s *ChiServer) home(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ACCRUAL.GOPHERMART"))
}

func (s *ChiServer) getOrderAccrual(w http.ResponseWriter, r *http.Request) {

	number := chi.URLParam(r, "orderNumRaw")
	if number == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	orderNum, err := strconv.Atoi(number)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// luna validation
	// https://ru.wikipedia.org/wiki/Алгоритм_Луна
	// https://goodcalculators.com/luhn-algorithm-calculator/?Num=18
	if ok := luhn.Valid(orderNum); !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// reading from db

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(""))
}
