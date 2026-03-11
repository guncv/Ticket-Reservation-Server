package log

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/shared"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	loggerInstance *logger
	once           sync.Once
)

type Logger interface {
	Error(ctx context.Context, args ...interface{})
	Debug(ctx context.Context, args ...interface{})
	Info(ctx context.Context, args ...interface{})
	Warn(ctx context.Context, args ...interface{})
}

type logger struct {
	*zap.SugaredLogger
}

func NewLogger(config *config.Config) Logger {
	return Initialize(config.AppConfig.AppEnv)
}

func Initialize(appEnv string) *logger {
	once.Do(func() {
		var baseLogger *zap.Logger
		var err error

		switch appEnv {
		case shared.AppEnvTest:
			baseLogger = zap.NewNop()
		case shared.AppEnvDev:
			config := zap.NewDevelopmentConfig()
			config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
			config.EncoderConfig.TimeKey = ""
			config.EncoderConfig.CallerKey = "caller"
			config.EncoderConfig.MessageKey = "msg"
			config.EncoderConfig.LevelKey = "level"
			config.EncoderConfig.ConsoleSeparator = " | "
			baseLogger, err = config.Build(zap.AddCaller())
			if err != nil {
				panic("failed to initialize zap logger: " + err.Error())
			}
		default:
			config := zap.NewProductionConfig()
			config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			baseLogger, err = config.Build(zap.AddCaller())
			if err != nil {
				panic("failed to initialize zap logger: " + err.Error())
			}
		}

		loggerInstance = &logger{baseLogger.Sugar()}
	})

	return loggerInstance
}

func (l *logger) Error(ctx context.Context, args ...interface{}) {
	loggerWithSkip := l.SugaredLogger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar()
	loc, _ := time.LoadLocation("Asia/Bangkok")
	timestamp := time.Now().In(loc).Format(time.RFC3339)
	loggerWithSkip.Errorf("%s | TimeStamp: %s", fmt.Sprint(args...), timestamp)
}

func (l *logger) Debug(ctx context.Context, args ...interface{}) {
	loggerWithSkip := l.SugaredLogger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar()
	loc, _ := time.LoadLocation("Asia/Bangkok")
	timestamp := time.Now().In(loc).Format(time.RFC3339)
	loggerWithSkip.Debugf("%s | TimeStamp: %s", fmt.Sprint(args...), timestamp)
}

func (l *logger) Warn(ctx context.Context, args ...interface{}) {
	loggerWithSkip := l.SugaredLogger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar()
	loc, _ := time.LoadLocation("Asia/Bangkok")
	timestamp := time.Now().In(loc).Format(time.RFC3339)
	loggerWithSkip.Warnf("%s | TimeStamp: %s", fmt.Sprint(args...), timestamp)
}

func (l *logger) Info(ctx context.Context, args ...interface{}) {
	loggerWithSkip := l.SugaredLogger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar()
	loc, _ := time.LoadLocation("Asia/Bangkok")
	timestamp := time.Now().In(loc).Format(time.RFC3339)
	loggerWithSkip.Infof("%s | TimeStamp: %s", fmt.Sprint(args...), timestamp)
}
