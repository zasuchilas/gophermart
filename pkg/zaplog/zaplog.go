package zaplog

import (
	"go.uber.org/zap"
	"runtime/debug"
)

func Initialize(level string, isProd bool) (*zap.Logger, error) {
	// parsing level
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	// setting the configuration
	var cfg zap.Config
	if isProd {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}
	cfg.Level = lvl

	// creating a logger based on the configuration
	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return zl, nil
}

func ServiceInfo(l *zap.Logger, title, appVersion string) {
	// get app module name
	buildInfo, ok := debug.ReadBuildInfo()
	var moduleName string
	if !ok {
		l.Error("Failed to read build info")
		moduleName = "-"
	} else {
		moduleName = buildInfo.Main.Path
	}

	// write data to the log
	l.Info(
		title,
		zap.String("name", moduleName),
		zap.String("version", appVersion),
	)
}
