package worker

import (
	"context"
	"fmt"
	"github.com/zasuchilas/gophermart/internal/accrual/config"
	"github.com/zasuchilas/gophermart/internal/accrual/logger"
	"github.com/zasuchilas/gophermart/internal/accrual/models"
	"github.com/zasuchilas/gophermart/internal/accrual/storage"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

type CalculateAccrualWorker struct {
	waitGroup *sync.WaitGroup
	store     storage.Storage
	timer     *time.Timer
	doneCh    chan struct{}
}

func New(store storage.Storage, wg *sync.WaitGroup) *CalculateAccrualWorker {
	wr := CalculateAccrualWorker{
		store:     store,
		timer:     time.NewTimer(config.WorkerPeriod),
		doneCh:    make(chan struct{}),
		waitGroup: wg,
	}
	return &wr
}

func (w *CalculateAccrualWorker) Start() {
loop:
	for {
		select {
		case <-w.doneCh:
			// time to closing
			break loop
		case <-w.timer.C:
			// time to working

			// TODO: generator -> worker pool
			//  - sharing access to shared data

			// getting goods
			goods, err := w.store.GetGoods(context.TODO())
			if err != nil {
				logger.Log.Info("error getting goods from db", zap.String("error", err.Error()))
				w.resetTimer()
				continue
			}
			if len(goods) == 0 {
				logger.Log.Debug("goods is empty")
				w.resetTimer()
				continue
			}

			// getting pack of orders
			orders, err := w.store.GetOrders(context.TODO())
			if err != nil {
				logger.Log.Info("error getting orders from db", zap.String("error", err.Error()))
				w.resetTimer()
				continue
			}
			if len(orders) == 0 {
				logger.Log.Debug("orders is empty (nothing to work)")
				w.resetTimer()
				continue
			}

			// processing every order
			w.processing(goods, orders)

			w.resetTimer()
		}
	}
}

func (w *CalculateAccrualWorker) Stop() {
	w.timer.Stop()
	w.doneCh <- struct{}{}
	w.waitGroup.Done()
}

func (w *CalculateAccrualWorker) resetTimer() {
	w.timer.Reset(config.WorkerPeriod)
}

func (w *CalculateAccrualWorker) processing(goods []*models.GoodsData, orders []*models.AccrualOrder) {
	for _, order := range orders {
		var (
			accrual float64
			err     error
		)

		// checking order
		goodsList := order.Receipt.Goods
		if goodsList == nil || len(goodsList) == 0 {
			err = w.store.UpdateOrder(context.TODO(), order.ID, storage.OrderStatusInvalid, 0)
			if err != nil {
				logger.Log.Info("error updating order",
					zap.String("order_num", order.OrderNum), zap.String("error", err.Error()))
			}
			continue
		}

		// calculating accrual
	loopPos:
		for _, position := range goodsList {
			ac, er := w.accrualOfReceiptPosition(&position, goods)
			if er != nil {
				err = er
				break loopPos
			}
			accrual += ac
		}
		if err != nil {
			err = w.store.UpdateOrder(context.TODO(), order.ID, storage.OrderStatusInvalid, 0)
			if err != nil {
				logger.Log.Info("error updating order",
					zap.String("order_num", order.OrderNum), zap.String("error", err.Error()))
			}
			continue
		}

		// updating order
		err = w.store.UpdateOrder(context.TODO(), order.ID, storage.OrderStatusProcessed, accrual)
		if err != nil {
			logger.Log.Info("error updating order",
				zap.String("order_num", order.OrderNum), zap.String("error", err.Error()))
		}
	}
}

func (w *CalculateAccrualWorker) accrualOfReceiptPosition(pos *models.GoodsPosition, goods []*models.GoodsData) (accrual float64, err error) {
loop:
	for _, good := range goods {
		if strings.Contains(pos.Description, good.Match) {
			reward, er := calculateAccrual(good.RewardType, good.Reward, pos.Price)
			if er != nil {
				err = er
				break loop
			}
			accrual += reward
		}
	}
	if err != nil {
		return 0, err
	}
	return accrual, nil
}

func calculateAccrual(rewardType string, reward float64, price float64) (accrual float64, err error) {
	if rewardType == "%" {
		return price / 100 * reward, nil
	}
	if rewardType == "pt" {
		return reward, nil
	}
	return 0, fmt.Errorf("unknown rewar type %s", rewardType)
}
