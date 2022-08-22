package logger

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	l       *zap.Logger
	pkgs    map[string]bool
	metrics bool

	LogCount *LogCounter
}

type LogCounter struct {
	rwMutex sync.RWMutex
	data    map[LogOrigin]uint32
}

type LogOrigin struct {
	File string
	Line int
}

func (lc *LogCounter) update(lo LogOrigin) {
	lc.rwMutex.Lock()
	lc.data[lo]++
	lc.rwMutex.Unlock()
}

func (lc *LogCounter) Lookup(key LogOrigin) (uint32, error) {
	lc.rwMutex.RLock()
	count, found := lc.data[key]
	lc.rwMutex.RUnlock()

	if !found {
		return 0, fmt.Errorf("LogCount key not found: %v", key)
	}

	return count, nil
}

func (lc *LogCounter) Dump() map[LogOrigin]uint32 {
	lc.rwMutex.RLock()
	dump := make(map[LogOrigin]uint32, len(lc.data))
	for k, v := range lc.data {
		dump[k] = v
	}
	lc.rwMutex.RUnlock()

	return dump
}

func newLogCounter() *LogCounter {
	return &LogCounter{
		rwMutex: sync.RWMutex{},
		data:    map[LogOrigin]uint32{},
	}
}

type Field = zap.Field
type Level = zapcore.Level

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
	Skip       = zap.Skip
	Binary     = zap.Binary
	Bool       = zap.Bool
	Boolp      = zap.Boolp
	ByteString = zap.ByteString

	Float64   = zap.Float64
	Float64p  = zap.Float64p
	Float32   = zap.Float32
	Float32p  = zap.Float32p
	Durationp = zap.Durationp

	Any = zap.Any
)

func getCallerInfo(skip int) (pkg, file string, line int) {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		panic("could not get runtime caller information")
	}

	funcName := runtime.FuncForPC(pc).Name()
	lastSlash := strings.LastIndexByte(funcName, '/')
	if lastSlash < 0 {
		lastSlash = 0
	}
	firstDot := strings.Index(funcName[lastSlash:], ".") + lastSlash
	pkg = funcName[:firstDot]

	return pkg, file, line
}

func (l *Logger) updateMetrics(file string, line int) {
	lo := LogOrigin{
		File: file,
		Line: line,
	}
	l.LogCount.update(lo)
}

func (l *Logger) checkEnabledAndUpdateMetrics() bool {
	pkg, file, line := getCallerInfo(2)

	// filter out not enbaled packages
	if len(l.pkgs) > 0 {
		if _, enabled := l.pkgs[pkg]; !enabled {
			return false
		}
	}

	if l.metrics {
		l.updateMetrics(file, line)
	}

	return true
}

func (l *Logger) Debug(msg string, fields ...Field) {
	if !l.checkEnabledAndUpdateMetrics() {
		return
	}

	l.l.Debug(msg, fields...)
}

func (l *Logger) Info(msg string, fields ...Field) {
	if !l.checkEnabledAndUpdateMetrics() {
		return
	}

	l.l.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	if !l.checkEnabledAndUpdateMetrics() {
		return
	}

	l.l.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	if !l.checkEnabledAndUpdateMetrics() {
		return
	}

	l.l.Error(msg, fields...)
}

func (l *Logger) DPanic(msg string, fields ...Field) {
	if !l.checkEnabledAndUpdateMetrics() {
		return
	}

	l.l.DPanic(msg, fields...)
}

func (l *Logger) Panic(msg string, fields ...Field) {
	if !l.checkEnabledAndUpdateMetrics() {
		return
	}

	l.l.Panic(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...Field) {
	if !l.checkEnabledAndUpdateMetrics() {
		return
	}

	l.l.Fatal(msg, fields...)
}

func (l *Logger) Sync() error {
	return l.l.Sync()
}

func New(w io.Writer, lvl Level, packages ...string) *Logger {
	if w == nil {
		panic("logger writer must be set")
	}

	var (
		cfg     zap.Config
		encoder zapcore.Encoder
		metrics bool
	)

	if lvl == DebugLevel {
		cfg = zap.NewDevelopmentConfig()
		encoder = zapcore.NewConsoleEncoder(cfg.EncoderConfig)
		metrics = true // enabled only for DebugLevel
	} else {
		cfg = zap.NewProductionConfig()
		encoder = zapcore.NewJSONEncoder(cfg.EncoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(w),
		zapcore.Level(lvl),
	)

	pkgs := map[string]bool{}
	for _, pkg := range packages {
		pkgs[pkg] = true
	}

	return &Logger{
		l:        zap.New(core),
		pkgs:     pkgs,
		metrics:  metrics,
		LogCount: newLogCounter(),
	}
}
