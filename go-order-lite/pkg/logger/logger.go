package logger

import "go.uber.org/zap"

var Log *zap.Logger

func Init(level string) error {
	cfg := zap.NewProductionConfig()
	logger, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = logger
	return nil
}
