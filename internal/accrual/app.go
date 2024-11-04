package accrual

import (
	"github.com/zasuchilas/gophermart/internal/accrual/config"
	"github.com/zasuchilas/gophermart/internal/accrual/logger"
	"github.com/zasuchilas/gophermart/internal/accrual/server"
	"github.com/zasuchilas/gophermart/internal/accrual/server/chisrv"
	"github.com/zasuchilas/gophermart/internal/accrual/storage"
	"github.com/zasuchilas/gophermart/internal/accrual/storage/pgstorage"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type App struct {
	AppName    string
	AppVersion string
	waitGroup  *sync.WaitGroup
	store      storage.Storage
	server     server.Server
}

func New() *App {
	config.ParseFlags()
	wg := &sync.WaitGroup{}

	return &App{
		AppName:    "accrual.gophermart",
		AppVersion: "0.0.0",
		waitGroup:  wg,
	}
}

func (a *App) Run() {
	logger.Init()
	logger.ServiceInfo("ACCRUAL.GOPHERMART (... service)", a.AppVersion)
	a.store = pgstorage.New()
	a.server = chisrv.New(a.store)
	a.waitGroup.Add(1)
	go a.server.Start()
	a.shutdown()
	a.waitGroup.Wait()
}

func (a *App) shutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-sigChan
		logger.Log.Info("The stop signal has been received", zap.String("signal", sig.String()))
		close(sigChan)

		a.store.Stop()
		a.server.Stop()

		logger.Log.Info("ACCRUAL.GOPHERMART service stopped")
		a.waitGroup.Done()
	}()
}
