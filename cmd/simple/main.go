package main

import (
	"fmt"
	"os"
	"strings"

	logger "github.com/geyslan/logger-poc/logger/simple"
	outerpackage "github.com/geyslan/logger-poc/outerpackagesimple"
)

func generateLogs(l *logger.Logger) {
	fmt.Println()
	fmt.Println(strings.Repeat("-", 5), "generateLogs()")

	l.Error("error message 1",
		logger.Bool("tested", true),
		logger.Float32("floated", 1.2),
	)
	l.Error("error message 2",
		logger.Bool("tested", false),
		logger.Float32("floated", 1.9),
	)
	l.Debug("debug message 1")
	l.Debug("debug message 2")
	l.Debug("debug message 3")
}

func generateLogsOuter() {
	fmt.Println()
	fmt.Println(strings.Repeat("-", 5), "generateLogsOuter()")

	outerpackage.OuterLog()
}

func dumpMetrics(l *logger.Logger) {
	fmt.Println()
	fmt.Println(strings.Repeat(">", 5), "Dump LogCount")

	logs := l.LogCount.Dump()
	fmt.Println("", strings.Repeat(">", 4), "Logs path count:", len(logs))
	for k, v := range logs {
		fmt.Printf("key: %+v - value: %v\n", k, v)
	}
}

func main() {
	generateLogs(logger.Default())
	dumpMetrics(logger.Default())

	generateLogsOuter()
	dumpMetrics(logger.Default())

	generateLogs(logger.Default())
	generateLogsOuter()
	dumpMetrics(logger.Default())

	if !logger.IsSetFromEnv() {
		fmt.Println()
		fmt.Println(strings.Repeat("*", 5), "Set (reset) new default logger")
		logger.Reset(
			&logger.LoggerConfig{
				Writer:  os.Stderr,
				Level:   logger.DebugLevel,
				Encoder: logger.NewJSONEncoder(logger.NewProductionConfig().EncoderConfig),
			},
		)
	}

	generateLogs(logger.Default())
	generateLogsOuter()
	generateLogs(logger.Default())
	dumpMetrics(logger.Default())
}
