package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Rhymond/go-money"
	"github.com/zasuchilas/gophermart/internal/common"
	"github.com/zasuchilas/gophermart/internal/gophermart/config"
	"github.com/zasuchilas/gophermart/internal/gophermart/logger"
	"github.com/zasuchilas/gophermart/internal/gophermart/models"
	"github.com/zasuchilas/gophermart/internal/gophermart/storage"
	"go.uber.org/zap"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type OrderEnrichWorker struct {
	waitGroup *sync.WaitGroup
	store     storage.Storage
	timer     *time.Timer
	doneCh    chan struct{}
	throttle  atomic.Bool
	poolSize  atomic.Int32
}

func New(store storage.Storage, wg *sync.WaitGroup) *OrderEnrichWorker {
	wr := OrderEnrichWorker{
		store:     store,
		timer:     time.NewTimer(config.WorkerPeriod),
		doneCh:    make(chan struct{}),
		waitGroup: wg,
	}
	wr.throttle.Store(false)
	wr.poolSize.Store(int32(config.WorkerPoolSize))
	return &wr
}

func (w *OrderEnrichWorker) Start() {
loop:
	for {
		select {
		case <-w.doneCh:
			// time to close
			break loop
		case <-w.timer.C:
			// time to work
			w.throttle.Store(false)

			// getting pack of order for processing
			orders, err := w.store.GetOrdersPack(context.TODO())
			if err != nil {
				logger.Log.Info("error getting orders from db", zap.String("error", err.Error()))
				w.resetTimer()
				continue
			}
			if len(orders) == 0 {
				logger.Log.Debug("orders is empty")
				w.resetTimer()
				continue
			}

			// creating channel for jobs
			jobCount := len(orders)
			jobs := make(chan *models.OrderRow, jobCount)

			// creating worker pool
			psize := int(w.poolSize.Load())
			for i := 0; i < psize; i++ {
				go w.workerProc(jobs)
			}

			// sending jobs
		workerJobsLoop:
			for i := 0; i < jobCount; i++ {
				if throttle := w.throttle.Load(); throttle {
					logger.Log.Info("throttle!")
					close(jobs)
					break workerJobsLoop
				}
				jobs <- orders[i]
			}

			close(jobs)
			w.resetTimer()

			// TODO: measure the execution time and calculate the current speed
			//  - increase the pool size if you do not receive a throttle error
		}
	}
}

func (w *OrderEnrichWorker) Stop() {
	w.timer.Stop()
	w.doneCh <- struct{}{}
	w.waitGroup.Done()
}

func (w *OrderEnrichWorker) resetTimer() {
	w.timer.Reset(config.WorkerPeriod)
}

func (w *OrderEnrichWorker) throttlePause() {
	w.throttle.Store(true)
	w.timer.Reset(10 * time.Second) // time.Minute
}

func (w *OrderEnrichWorker) workerProc(jobs <-chan *models.OrderRow) {
	for order := range jobs {
		if throttle := w.throttle.Load(); throttle {
			continue
		}
		// do request
		// GET /api/orders/{number}
		u := fmt.Sprintf("http://%s/api/orders/%s", config.AccrualSystemAddress, order.OrderNum)
		response, err := http.Get(u)
		if err != nil {
			logger.Log.Info("getting error during request", zap.String("error", err.Error()))
			continue
		}
		if response.StatusCode == http.StatusTooManyRequests {
			w.throttlePause()
			continue
		}
		if response.StatusCode == http.StatusNoContent {
			continue
		}

		// decoding response
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			logger.Log.Info("cannot read response body", zap.String("error", err.Error()))
			continue
		}
		var resp models.OrderStateResponse
		err = json.Unmarshal(body, &resp)
		if err != nil {
			logger.Log.Info("cannot decode response JSON body",
				zap.String("error", err.Error()), zap.String("order_num", order.OrderNum))
			continue // ??
		}

		// write result into db
		if order.Status == resp.Status {
			continue
		}
		err = w.store.UpdateOrder(context.TODO(), order.UserID, order.ID, resp.Status, money.NewFromFloat(resp.Accrual, common.Currency))
		if err != nil {
			logger.Log.Info("error updating order data in db", zap.String("error", err.Error()))
			continue
		}
	}
}
