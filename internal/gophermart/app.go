package gophermart

import (
	"github.com/zasuchilas/gophermart/internal/gophermart/config"
	"github.com/zasuchilas/gophermart/internal/gophermart/logger"
	"github.com/zasuchilas/gophermart/internal/gophermart/server"
	"github.com/zasuchilas/gophermart/internal/gophermart/server/chisrv"
	"github.com/zasuchilas/gophermart/internal/gophermart/storage"
	"github.com/zasuchilas/gophermart/internal/gophermart/storage/pgstorage"
	"github.com/zasuchilas/gophermart/internal/gophermart/worker"
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
	worker     *worker.OrderEnrichWorker
}

func New() *App {
	config.ParseFlags()
	wg := &sync.WaitGroup{}

	return &App{
		AppName:    "gophermart",
		AppVersion: "0.0.0",
		waitGroup:  wg,
	}
}

func (a *App) Run() {
	logger.Init()
	logger.ServiceInfo("GOPHERMART (... service)", a.AppVersion)
	a.store = pgstorage.New()

	chisrv.InitJWT()
	a.server = chisrv.New(a.store, a.waitGroup)
	a.waitGroup.Add(1)
	go a.server.Start()

	a.worker = worker.New(a.store, a.waitGroup)
	a.waitGroup.Add(1)
	go a.worker.Start()

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

		a.worker.Stop()
		a.store.Stop()
		a.server.Stop()

		logger.Log.Info("GOPHERMART service stopped")
	}()
}
