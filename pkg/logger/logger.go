package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func Init() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	log, err = config.Build()
	if err != nil {
		panic(err)
	}
}

func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}
