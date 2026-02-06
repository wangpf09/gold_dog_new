package logger

import "go.uber.org/zap"

func Debug(msg string, fields ...zap.Field) {
	log().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	log().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	log().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	log().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	log().Fatal(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	log().Panic(msg, fields...)
}

func Debugf(format string, args ...interface{}) {
	sugar().Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	sugar().Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	sugar().Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	sugar().Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	sugar().Fatalf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	sugar().Panicf(format, args...)
}

func Debugln(args ...interface{}) {
	sugar().Debugln(args...)
}

func Infoln(args ...interface{}) {
	sugar().Infoln(args...)
}

func Warnln(args ...interface{}) {
	sugar().Warnln(args...)
}

func Errorln(args ...interface{}) {
	sugar().Errorln(args...)
}

func Fatalln(args ...interface{}) {
	sugar().Fatalln(args...)
}

func Panicln(args ...interface{}) {
	sugar().Panicln(args...)
}
