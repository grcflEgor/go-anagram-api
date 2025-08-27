package logger

import (
	"log"

	"go.uber.org/zap"
)


var AppLogger *zap.Logger

func InitLogger() {
	var err error

	AppLogger, err = zap.NewProduction()
	if err != nil {
		log.Fatalf("cant init zap logger: %v", err)
	}
}