package outerpackage

import logger "github.com/geyslan/logger-poc/logger/simple"

func OuterLog() {
	// Log using same execution time logger (package level)

	logger.Debug("1 Debbuging from outer package")
	logger.Debug("2 Debbuging from outer package")
	logger.Error("1 Error from outer package")
	logger.Error("2 Error from outer package")
}
