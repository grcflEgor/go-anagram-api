package logger

import (

	"go.uber.org/zap"
)


var AppLogger *zap.Logger

func InitLogger() {
	var err error

	AppLogger, err = zap.NewProduction()
	if err != nil {
		AppLogger.Fatal("cant init zap logger", zap.Error(err))
	}
}