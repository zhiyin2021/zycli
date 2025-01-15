package logger

import (
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	sugar       = NewSugar()
	atomicLevel = zap.NewAtomicLevel()
)

func NewLogger() *zap.Logger {
	// 创建 Zap Core
	atomicLevel.SetLevel(zapcore.InfoLevel)
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = EncodeTime
	config.TimeKey = "tm"
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		zapcore.AddSync(zapcore.Lock(os.Stdout)),
		atomicLevel, // 日志级别
	)
	logger := zap.New(core)
	return logger
}

func NewSugar() *zap.SugaredLogger {
	return NewLogger().Sugar()
}

func Infow(msg string, keysAndValues ...any) {
	sugar.Infow(msg, keysAndValues...)
}
func Infof(msg string, args ...any) {
	sugar.Infof(msg, args...)
}
func Infoln(msg ...any) {
	sugar.Infoln(msg...)
}
func Println(msg ...any) {
	sugar.Infoln(msg...)
}
func Errorw(msg string, keysAndValues ...any) {
	sugar.Errorw(msg, keysAndValues...)
}
func Errorf(msg string, args ...any) {
	sugar.Errorf(msg, args...)
}

func Errorln(msg ...any) {
	sugar.Errorln(msg...)
}
func Warnw(msg string, keysAndValues ...any) {
	sugar.Warnw(msg, keysAndValues...)
}
func Warnf(msg string, args ...any) {
	sugar.Warnf(msg, args...)
}

func Warnln(msg ...any) {
	sugar.Warnln(msg...)
}

func Debugw(msg string, keysAndValues ...any) {
	sugar.Debugw(msg, keysAndValues...)
}
func Debugf(msg string, args ...any) {
	sugar.Debugf(msg, args...)
}
func Debugln(msg ...any) {
	sugar.Debugln(msg...)
}

func Sync() {
	sugar.Sync()
}

func With(fields ...any) *zap.SugaredLogger {
	return sugar.With(fields...)
}

func EncodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("150405.000"))
}

func SetLogger(w io.Writer) {
	if sugar != nil {
		sugar.Sync()
	}

	atomicLevel.SetLevel(zapcore.InfoLevel)
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = EncodeTime
	config.TimeKey = "tm"
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout), // 输出到标准输出
			zapcore.AddSync(w),         // 输出到文件
		),
		zapcore.InfoLevel,
	)
	logger := zap.New(core)
	sugar = logger.Sugar()
}

func SetLevel(level zapcore.Level) {
	atomicLevel.SetLevel(level)
}
