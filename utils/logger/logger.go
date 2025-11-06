package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func NewLogger(env string) (*zap.Logger, error) {
	var zapLogger *zap.Logger
	var err error
	if env == "production" || env == "LIVE" {
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		zapLogger = zap.New(zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.Lock(os.Stdout),
			zapcore.DebugLevel,
		), zap.AddCaller(), zap.AddCallerSkip(1))
	} else {
		zapLogger, err = zap.NewDevelopment()
	}
	if err != nil {
		return nil, err
	}

	return zapLogger, nil
}
