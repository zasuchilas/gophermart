package gophermart

import (
	"github.com/zasuchilas/gophermart/internal/gophermart/config"
	"github.com/zasuchilas/gophermart/internal/gophermart/logger"
	"github.com/zasuchilas/gophermart/internal/gophermart/storage"
	"github.com/zasuchilas/gophermart/internal/gophermart/storage/pgstorage"
	"sync"
)

type App struct {
	AppName    string
	AppVersion string
	waitGroup  *sync.WaitGroup
	store      storage.Storage
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

}
