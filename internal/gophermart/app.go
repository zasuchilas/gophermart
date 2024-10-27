package gophermart

import "sync"

type App struct {
	AppName    string
	AppVersion string
	waitGroup  *sync.WaitGroup
}

func New() *App {
	parseFlags()
	wg := &sync.WaitGroup{}

	return &App{
		AppName:    "gophermart",
		AppVersion: "0.0.0",
		waitGroup:  wg,
	}
}

func (a *App) Run() {

}
