package logger

import (
	"github.com/zasuchilas/gophermart/internal/gophermart/config"
	"github.com/zasuchilas/gophermart/pkg/zaplog"
	"go.uber.org/zap"
	"log"
)

var Log *zap.Logger

func Init() {
	l, err := zaplog.Initialize(config.LogLevel, config.EnvType == "production")
	if err != nil {
		log.Fatal(err.Error())
	}
	Log = l
}

func ServiceInfo(title, appVersion string) {
	zaplog.ServiceInfo(Log, title, appVersion)

	Log.Info("Config",
		zap.String("RunAddress", config.RunAddress),
		zap.String("AccrualSystemAddress", config.AccrualSystemAddress),
	)
}
