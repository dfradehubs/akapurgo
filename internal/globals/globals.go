package globals

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

func GetLogger(logLevel string, disableTrace bool) (logger *zap.SugaredLogger, err error) {
	parsedLogLevel, err := zap.ParseAtomicLevel(logLevel)
	if err != nil {
		return logger, err
	}

	// Initialize the logger
	loggerConfig := zap.NewProductionConfig()
	if disableTrace {
		loggerConfig.DisableStacktrace = true
		loggerConfig.DisableCaller = true
	}

	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	loggerConfig.Level.SetLevel(parsedLogLevel.Level())

	// Configure the logger
	loggerObj, err := loggerConfig.Build()
	if err != nil {
		return logger, err
	}

	return loggerObj.Sugar(), nil
}
