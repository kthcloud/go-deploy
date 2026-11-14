package log

import (
	"sync"

	"github.com/kthcloud/go-deploy/models/mode"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	loggerMu sync.RWMutex

	// Logger is the global fallback logger used by legacy callers.
	// It defaults to a no-op logger so unit tests and early callers never panic.
	Logger    = zap.NewNop().Sugar()
	LoggerMap = make(map[string]*zap.SugaredLogger)
)

var (
	Bold  = "\033[1m"
	Reset = "\033[0m"

	Orange = "\033[38;5;208m"
	Grey   = "\033[90m"

	runMode       = mode.Dev
	defaultLogger = "default"
)

// New constructs a zap.Logger configured for the given run mode.
func New(m string) (*zap.Logger, error) {
	if m == mode.Prod {
		return zap.NewProduction(zap.WithCaller(false))
	}
	return zap.NewDevelopment(zap.WithCaller(false))
}

// SetupLogger initializes the legacy global logger for the given run mode.
func SetupLogger(m string) error {
	base, err := New(m)
	if err != nil {
		return err
	}

	loggerMu.Lock()
	runMode = m
	LoggerMap = make(map[string]*zap.SugaredLogger)
	if m == mode.Prod {
		Bold = ""
		Reset = ""
		Orange = ""
		Grey = ""
	}
	Logger = base.Sugar()
	LoggerMap[defaultLogger] = Logger
	loggerMu.Unlock()

	return nil
}

// Get returns a named legacy SugaredLogger, creating it safely on first access.
func Get(name string) *zap.SugaredLogger {
	loggerMu.RLock()
	if sugaredLogger, ok := LoggerMap[name]; ok {
		loggerMu.RUnlock()
		return sugaredLogger
	}
	m := runMode
	loggerMu.RUnlock()

	base, err := New(m)
	if err != nil {
		base = zap.NewNop()
	}

	var sugared *zap.SugaredLogger
	if m != mode.Prod && name == defaultLogger {
		sugared = base.Sugar()
	} else {
		sugared = base.Sugar().Named(name)
	}

	loggerMu.Lock()
	if existing, ok := LoggerMap[name]; ok {
		loggerMu.Unlock()
		return existing
	}
	LoggerMap[name] = sugared
	loggerMu.Unlock()

	return sugared
}

// Sync flushes any buffered log entries from the global logger.
func Sync() {
	loggerMu.RLock()
	l := Logger
	loggerMu.RUnlock()
	if l != nil {
		_ = l.Sync()
	}
}

// Logln logs a message at provided level.
// Spaces are always added between arguments.
func Logln(lvl zapcore.Level, args ...any) {
	Logger.Logln(lvl, args...)
}

// Debugln logs a message at [DebugLevel].
// Spaces are always added between arguments.
func Debugln(args ...any) {
	Logger.Debugln(args...)
}

// Println logs a message at [InfoLevel].
// Spaces are always added between arguments.
// This is an alias for Infoln.
func Println(args ...any) {
	Logger.Infoln(args...)
}

// Infoln logs a message at [InfoLevel].
// Spaces are always added between arguments.
func Infoln(args ...any) {
	Logger.Infoln(args...)
}

// Warnln logs a message at [WarnLevel].
// Spaces are always added between arguments.
func Warnln(args ...any) {
	Logger.Warnln(args...)
}

// Errorln logs a message at [ErrorLevel].
// Spaces are always added between arguments.
func Errorln(args ...any) {
	Logger.Errorln(args...)
}

// Fatalln logs a message at [FatalLevel] and then calls os.Exit(1).
// Spaces are always added between arguments.
func Fatalln(args ...any) {
	Logger.Fatalln(args...)
}

// Logf formats the message according to the format specifier
// and logs it at provided level.
func Logf(lvl zapcore.Level, template string, args ...any) {
	Logger.Logf(lvl, template, args...)
}

// Debugf formats the message according to the format specifier
// and logs it at [DebugLevel].
func Debugf(template string, args ...any) {
	Logger.Debugf(template, args...)
}

// Printf formats the message according to the format specifier
// and logs it at [InfoLevel].
// This is an alias for Infof.
func Printf(template string, args ...any) {
	Logger.Infof(template, args...)
}

// Infof formats the message according to the format specifier
// and logs it at [InfoLevel].
func Infof(template string, args ...any) {
	Logger.Infof(template, args...)
}

// Warnf formats the message according to the format specifier
// and logs it at [WarnLevel].
func Warnf(template string, args ...any) {
	Logger.Warnf(template, args...)
}

// Errorf formats the message according to the format specifier
// and logs it at [ErrorLevel].
func Errorf(template string, args ...any) {
	Logger.Errorf(template, args...)
}

// Fatalf formats the message according to the format specifier
// and logs it at [FatalLevel].
func Fatalf(template string, args ...any) {
	Logger.Fatalf(template, args...)
}
