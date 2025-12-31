package log

import (
	"fmt"
	"os"
	"time"

	"github.com/go-kratos/kratos-layout/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type zapWrapper func(level log.Level, keyvals ...interface{}) error

func (f zapWrapper) Log(level log.Level, keyvals ...interface{}) error {
	return f(level, keyvals...)
}

func NewZapLogger(env conf.Environment, logPath string, maxSize, maxAge, level, maxBackups int32) log.Logger {
	logLevel := zapcore.Level(level)

	fileRotate := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    int(maxSize),
		MaxAge:     int(maxAge),
		MaxBackups: int(maxBackups),
		LocalTime:  true,
	}

	WriterSyncer := []zapcore.WriteSyncer{
		zapcore.AddSync(fileRotate),
	}

	zapOpts := make([]zap.Option, 0)
	zapOpts = append(zapOpts, zap.AddStacktrace(
		zap.NewAtomicLevelAt(zapcore.ErrorLevel)),
		zap.AddCaller(),
		zap.AddCallerSkip(4),
	)

	if env != conf.Environment_PROD {
		logLevel = zapcore.DebugLevel
		WriterSyncer = append(WriterSyncer, zapcore.AddSync(os.Stdout))
		zapOpts = append(zapOpts, zap.Development())
	}

	coreInfo := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:       "ts",
			LevelKey:      "level",
			NameKey:       "name",
			CallerKey:     "caller",
			MessageKey:    "message",
			StacktraceKey: "trace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.CapitalLevelEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
			},
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}),
		zapcore.NewMultiWriteSyncer(WriterSyncer...),
		zap.NewAtomicLevelAt(logLevel),
	)

	core := zapcore.NewTee(coreInfo)

	zapLogger := zap.New(core, zapOpts...)

	return zapWrapper(func(level log.Level, keyvals ...interface{}) error {
		switch level {
		case log.LevelDebug:
			logLevel = zap.DebugLevel
		case log.LevelInfo:
			logLevel = zap.InfoLevel
		case log.LevelWarn:
			logLevel = zap.WarnLevel
		case log.LevelError:
			logLevel = zap.ErrorLevel
		case log.LevelFatal:
			logLevel = zap.FatalLevel
		}
		var fields []zap.Field
		for i := 0; i < len(keyvals); i += 2 {
			fields = append(fields, zap.String(fmt.Sprintf("%v", keyvals[i]), fmt.Sprintf("%v", keyvals[i+1])))
		}
		zapLogger.Log(logLevel, "log", fields...)
		return nil
	})
}
