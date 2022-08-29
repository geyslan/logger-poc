package logger

import (
	"io"
	"os"
	"runtime"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	LOGGER_LVL = "LOGGER_LVL"
)

type (
	Field = zap.Field
	Level = zapcore.Level
)

const (
	DebugLevel  Level = zap.DebugLevel
	InfoLevel   Level = zap.InfoLevel
	WarnLevel   Level = zap.WarnLevel
	ErrorLevel  Level = zap.ErrorLevel
	DPanicLevel Level = zap.DPanicLevel
	PanicLevel  Level = zap.PanicLevel
	FatalLevel  Level = zap.FatalLevel
)

var (
	// TODO: envelope and rename all zap fields
	Any        = zap.Any
	Skip       = zap.Skip
	Binary     = zap.Binary
	Bool       = zap.Bool
	Boolp      = zap.Boolp
	ByteString = zap.ByteString

	Uint32  = zap.Uint32
	Uint32p = zap.Uint32p
	Uint32s = zap.Uint32s

	Float64    = zap.Float64
	Float64p   = zap.Float64p
	Float32    = zap.Float32
	Float32p   = zap.Float32p
	ErrorField = zap.Error

	Durationp = zap.Durationp

	String  = zap.String
	Stringp = zap.Stringp
)

type Config = zap.Config

var (
	NewDevelopmentConfig = zap.NewDevelopmentConfig
	NewProductionConfig  = zap.NewProductionConfig
)

type Encoder = zapcore.Encoder

var (
	NewConsoleEncoder = zapcore.NewConsoleEncoder
	NewJSONEncoder    = zapcore.NewJSONEncoder
)

// Logger struct
type Logger struct {
	l   *zap.Logger
	cfg *LoggerConfig

	LogCount *LogCounter // updated only on debug level
}

// NewLogger function
func NewLogger(cfg *LoggerConfig) *Logger {
	if cfg == nil {
		panic("LoggerConfig cannot be nil")
	}

	core := zapcore.NewCore(
		cfg.Encoder,
		zapcore.AddSync(cfg.Writer),
		zapcore.Level(cfg.Level),
	)

	return &Logger{
		l:        zap.New(core),
		cfg:      cfg,
		LogCount: newLogCounter(),
	}
}

type LoggerConfig struct {
	Writer  io.Writer
	Level   Level
	Encoder Encoder
}

func NewDefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		os.Stderr,
		InfoLevel,
		NewJSONEncoder(NewProductionConfig().EncoderConfig),
	}
}

type LogOrigin struct {
	File string
	Line int
}

type LogCounter struct {
	rwMutex sync.RWMutex
	data    map[LogOrigin]uint32
}

func (lc *LogCounter) update(lo LogOrigin) (new bool) {
	lc.rwMutex.Lock()
	defer lc.rwMutex.Unlock()
	_, found := lc.data[lo]
	lc.data[lo]++

	return !found
}

func (lc *LogCounter) Lookup(key LogOrigin) (count uint32, found bool) {
	lc.rwMutex.RLock()
	defer lc.rwMutex.RUnlock()
	count, found = lc.data[key]

	return
}

func (lc *LogCounter) Dump() map[LogOrigin]uint32 {
	lc.rwMutex.RLock()
	defer lc.rwMutex.RUnlock()
	dump := make(map[LogOrigin]uint32, len(lc.data))
	for k, v := range lc.data {
		dump[k] = v
	}

	return dump
}

func newLogCounter() *LogCounter {
	return &LogCounter{
		rwMutex: sync.RWMutex{},
		data:    map[LogOrigin]uint32{},
	}
}

func getCallerInfo(skip int) (pkg, file string, line int) {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		panic("could not get runtime caller information")
	}

	funcName := runtime.FuncForPC(pc).Name()
	lastSlash := strings.LastIndexByte(funcName, '/')
	if lastSlash < 0 {
		lastSlash = 0
	} else {
		lastSlash++
	}
	lastDot := strings.LastIndexByte(funcName, '.')
	pkg = funcName[lastSlash:lastDot]
	// check if it's from a receiver
	if possibleLastDot := strings.LastIndexByte(pkg, '.'); possibleLastDot != -1 {
		pkg = pkg[0:possibleLastDot]
	}

	return pkg, file, line
}

func (l *Logger) updateCounter(file string, line int) (new bool) {
	return l.LogCount.update(LogOrigin{
		File: file,
		Line: line,
	})
}

// Log functions

// Debug
func debug(skip int, l *Logger, msg string, fields ...Field) {
	if l.cfg.Level > DebugLevel {
		return
	}

	_, file, line := getCallerInfo(skip + 1)
	if new := l.updateCounter(file, line); new {
		l.l.Debug(msg, fields...)
	}
}

func Debug(msg string, fields ...Field) {
	debug(1, defaultLogger, msg, fields...)
}

func (l *Logger) Debug(msg string, fields ...Field) {
	debug(1, l, msg, fields...)
}

// Error
func err(skip int, l *Logger, msg string, fields ...Field) {
	_, file, line := getCallerInfo(skip + 1)
	if new := l.updateCounter(file, line); new {
		l.l.Error(msg, fields...)
	}
}

func Error(msg string, fields ...Field) {
	err(1, defaultLogger, msg, fields...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	err(1, l, msg, fields...)
}

// Fatal
func fatal(skip int, l *Logger, msg string, fields ...Field) {
	_, file, line := getCallerInfo(skip + 1)
	if new := l.updateCounter(file, line); new {
		l.l.Fatal(msg, fields...)
	}
}

func Fatal(msg string, fields ...Field) {
	fatal(1, defaultLogger, msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...Field) {
	fatal(1, l, msg, fields...)
}

// Sync
func (l *Logger) Sync() error {
	return l.l.Sync()
}

func getLoggerLevelFromEnv() Level {
	lvlEnv := os.Getenv(LOGGER_LVL)

	setFromEnv = true

	switch lvlEnv {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "dpanic":
		return DPanicLevel
	case "panic":
		return PanicLevel
	case "fatal":
		return FatalLevel
	default:
		setFromEnv = false
		return InfoLevel
	}
}

// Package level default Logger
var defaultLogger *Logger

var setFromEnv bool

func IsSetFromEnv() bool {
	return setFromEnv
}

func init() {
	lvl := getLoggerLevelFromEnv()
	var enc Encoder

	if lvl == DebugLevel {
		enc = NewJSONEncoder(zap.NewDevelopmentEncoderConfig())
	} else {
		enc = NewJSONEncoder(zap.NewProductionEncoderConfig())
	}

	defaultLogger = NewLogger(
		&LoggerConfig{
			os.Stderr,
			lvl,
			enc,
		},
	)
}

func Default() *Logger {
	return defaultLogger
}

// Reset resets the package level default logger using given config.
// It's not thread safe so if required use it always at the beggining.
// Be assured that log count of the current default logger will be lost.
func Reset(cfg *LoggerConfig) {
	if cfg == nil {
		panic("LoggerConfig cannot be nil")
	}

	defaultLogger = NewLogger(cfg)
}
