package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides global logging instance
var Logger *zap.SugaredLogger

// InitLogger initializes logging instance
func InitLogger() {
	var loggingConfig zap.Config

	loggingConfig = zap.NewProductionConfig()
	// Remove this to make it JSON
	loggingConfig.Encoding = "console"

	loggingConfig.OutputPaths = append(loggingConfig.OutputPaths, "debug.log")
	loggingConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	loggingConfig.DisableCaller = true
	loggingConfig.DisableStacktrace = true

	lgConf, err := loggingConfig.Build()
	if err != nil {
		log.Fatal("Unable to build logger, error: ", err.Error())
	}
	defer lgConf.Sync()

	Logger = lgConf.Sugar()
	defer Logger.Sync()

	// Redirects logs made to console to our logger.
	// Ref: https://pkg.go.dev/go.uber.org/zap?tab=doc#RedirectStdLog
	// zap.RedirectStdLog(Logger.Desugar())
}
