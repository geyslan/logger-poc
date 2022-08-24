package outerpackage

import logger "github.com/geyslan/logger-poc/logger/zap"

type empty struct{}

func (e empty) fun(l *logger.Logger) {
	l.Debug("4 Debbuging from outer package")
}

func Log(l *logger.Logger) {
	l.Debug("3 Debbuging from outer package")
	empty{}.fun(l)
}
