package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger
var atom zap.AtomicLevel

func init() {
	atom = zap.NewAtomicLevel()
	atom.SetLevel(zapcore.InfoLevel)

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = ""

	l := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	))
	defer l.Sync()

	log = l
}

func SetDebug() {
	atom.SetLevel(zapcore.DebugLevel)
}

func Error(err error) {
	defer log.Sync()
	sugar := log.Sugar()
	sugar.Error(err)
}

func Warnf(template string, args ...interface{}) {
	defer log.Sync()
	sugar := log.Sugar()
	sugar.Warnf(template, args)
}

func Info(msg string, fields ...zap.Field) {
	defer log.Sync()
	sugar := log.Sugar()
	sugar.Info(msg, fields)
}

func Infof(template string, args ...interface{}) {
	defer log.Sync()
	sugar := log.Sugar()
	sugar.Infof(template, args)
}

func Debug(msg string, fields ...zap.Field) {
	defer log.Sync()
	sugar := log.Sugar()
	sugar.Debug(msg, fields)
}

func Debugf(template string, args ...interface{}) {
	defer log.Sync()
	sugar := log.Sugar()
	sugar.Debugf(template, args)
}
