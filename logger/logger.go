package logger

import (
	"io"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

type Level log.Level

type Logger struct {
	pkgs    map[string]bool
	modules uint64
}

func Init() *Logger {
	return &Logger{
		pkgs:    map[string]bool{},
		modules: 0,
	}
}

func getCallerPackage(skip int) string {
	pc, _, _, _ := runtime.Caller(skip + 1)
	funcName := runtime.FuncForPC(pc).Name()

	lastSlash := strings.LastIndexByte(funcName, '/')
	if lastSlash < 0 {
		lastSlash = 0
	}
	firstDot := strings.Index(funcName[lastSlash:], ".") + lastSlash

	return funcName[:firstDot]
}

func (l Logger) isPackageSet() bool {
	if len(l.pkgs) == 0 {
		return true
	}

	callerPackage := getCallerPackage(2)
	//fmt.Println("package", callerPackage)

	_, found := l.pkgs[callerPackage]
	return found
}

func (l *Logger) WarnfPkg(format string, args ...interface{}) {
	if !l.isPackageSet() {
		return
	}

	log.Warnf(format, args...)
}

func (l Logger) isModuleSet(moduleBitFlag uint64) bool {
	if l.modules == 0 {
		return true
	}

	return l.modules&moduleBitFlag != 0
}

func (l *Logger) WarnfModule(moduleBitFlag uint64, format string, args ...interface{}) {
	if !l.isModuleSet(moduleBitFlag) {
		return
	}

	log.Warnf(format, args...)
}

func (l *Logger) ConfigureMod(output io.Writer, modules uint64) {
	log.SetOutput(io.Discard)

	log.AddHook(&writer.Hook{
		Writer: output,
		LogLevels: []log.Level{
			log.WarnLevel,
			log.ErrorLevel,
			log.FatalLevel,
			log.PanicLevel,
		},
	})

	l.modules = modules
}

func (l *Logger) ConfigurePkg(output io.Writer, pkgs ...string) {
	log.SetOutput(io.Discard)

	log.AddHook(&writer.Hook{
		Writer: output,
		LogLevels: []log.Level{
			log.WarnLevel,
			log.ErrorLevel,
			log.FatalLevel,
			log.PanicLevel,
		},
	})

	if len(pkgs) > 0 {
		for _, pkg := range pkgs {
			l.pkgs[pkg] = true
		}
	}
}
