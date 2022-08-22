package main

import (
	"fmt"
	"os"
	"strings"

	logger "github.com/geyslan/logger-poc/logger/zap"
)

func generateLogs(log *logger.Logger) {
	fmt.Println()
	fmt.Println(strings.Repeat("-", 5), "generateLogs()")
	log.Warn("warn message 1",
		logger.Bool("tested", false),
		logger.ByteString("bytestring", []byte{65, 66}),
	)
	log.Warn("warn message 2",
		logger.Bool("tested", true),
		logger.Float32("floated", 1.2),
	)
	log.Error("error message 1",
		logger.Bool("tested", true),
		logger.Float32("floated", 1.2),
	)
	log.DPanic("dpanic message 1",
		logger.Bool("tested", true),
	)
	log.Debug("debug message 1")
}

func dumpMetrics(log *logger.Logger) {
	fmt.Println()
	fmt.Println(strings.Repeat(">", 5), "Dump LogCount")
	logs := log.LogCount.Dump()
	for k, v := range logs {
		fmt.Printf("key: %+v - value: %v\n", k, v)
	}
}

func main() {
	debugLogger := logger.New(os.Stdout, logger.DebugLevel)
	generateLogs(debugLogger)
	generateLogs(debugLogger)
	dumpMetrics(debugLogger)

	infoLogger := logger.New(os.Stdout, logger.InfoLevel)
	generateLogs(infoLogger)
}
