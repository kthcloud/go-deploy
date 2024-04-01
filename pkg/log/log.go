package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger      *zap.SugaredLogger
	AsyncLogger *zap.SugaredLogger
)

var (
	Bold  = "\033[1m"
	Reset = "\033[0m"

	Orange = "\033[38;5;208m"
	Grey   = "\033[90m"
)

func init() {
	development := true

	if development {
		l := zap.Must(zap.NewDevelopment(zap.WithCaller(false)))
		Logger = l.Sugar()
	} else {
		l := zap.Must(zap.NewProduction(zap.WithCaller(false)))
		Logger = l.Sugar()

		Bold = ""
		Reset = ""
		Orange = ""
		Grey = ""
	}
}

// Logln logs a message at provided level.
// Spaces are always added between arguments.
func Logln(lvl zapcore.Level, args ...interface{}) {
	Logger.Logln(lvl, args...)
}

// Debugln logs a message at [DebugLevel].
// Spaces are always added between arguments.
func Debugln(args ...interface{}) {
	Logger.Debugln(args...)
}

// Println logs a message at [InfoLevel].
// Spaces are always added between arguments.
// This is an alias for Infoln.
func Println(args ...interface{}) {
	Logger.Infoln(args...)
}

// Infoln logs a message at [InfoLevel].
// Spaces are always added between arguments.
func Infoln(args ...interface{}) {
	Logger.Infoln(args...)
}

// Warnln logs a message at [WarnLevel].
// Spaces are always added between arguments.
func Warnln(args ...interface{}) {
	Logger.Warnln(args...)
}

// Errorln logs a message at [ErrorLevel].
// Spaces are always added between arguments.
func Errorln(args ...interface{}) {
	Logger.Errorln(args...)
}

// Fatalln logs a message at [FatalLevel] and then calls os.Exit(1).
// Spaces are always added between arguments.
func Fatalln(args ...interface{}) {
	Logger.Fatalln(args...)
}

// Logf formats the message according to the format specifier
// and logs it at provided level.
func Logf(lvl zapcore.Level, template string, args ...interface{}) {
	Logger.Logf(lvl, template, args...)
}

// Debugf formats the message according to the format specifier
// and logs it at [DebugLevel].
func Debugf(template string, args ...interface{}) {
	Logger.Debugf(template, args...)
}

// Printf formats the message according to the format specifier
// and logs it at [InfoLevel].
// This is an alias for Infof.
func Printf(template string, args ...interface{}) {
	Logger.Infof(template, args...)
}

// Infof formats the message according to the format specifier
// and logs it at [InfoLevel].
func Infof(template string, args ...interface{}) {
	Logger.Infof(template, args...)
}

// Warnf formats the message according to the format specifier
// and logs it at [WarnLevel].
func Warnf(template string, args ...interface{}) {
	Logger.Warnf(template, args...)
}

// Errorf formats the message according to the format specifier
// and logs it at [ErrorLevel].
func Errorf(template string, args ...interface{}) {
	Logger.Errorf(template, args...)
}

// Fatalf formats the message according to the format specifier
// and logs it at [FatalLevel].
func Fatalf(template string, args ...interface{}) {
	Logger.Fatalf(template, args...)
}
