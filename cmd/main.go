package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/geyslan/logger-poc/logger"
)

const (
	EnrichModule uint64 = 1 << iota
	EBPFModule
)

type AnyWriter struct {
	LogChannel chan string
}

func InitAnyWriter() *AnyWriter {
	return &AnyWriter{
		LogChannel: make(chan string, 10),
	}
}

func (aw *AnyWriter) Write(message []byte) (n int, err error) {
	errMsg := string(message)
	aw.LogChannel <- errMsg

	return len(errMsg), nil
}

func generateModLogs(log *logger.Logger) {
	log.WarnfModule(EBPFModule, "%d %s", 2, "test", errors.New("new warn").Error())
	log.WarnfModule(EBPFModule, "%d %s", 2, "test", errors.New("new warn").Error())
	log.WarnfModule(EBPFModule, "%d %s", 2, "test", errors.New("new warn").Error())
}

func generatePkgLogs(log *logger.Logger) {
	log.WarnfPkg("%d %s", 1, "test", errors.New("new warn").Error())
	log.WarnfPkg("%d %s", 2, "test", errors.New("new warn").Error())
	log.WarnfPkg("%d %s", 3, "test", errors.New("new error").Error())
}

func main() {
	anyWriter := InitAnyWriter()

	logMod := logger.Init()
	logMod.ConfigureMod(anyWriter, EBPFModule|EnrichModule)

	go generateModLogs(logMod)
	fmt.Println(strings.Repeat("=", 5), "by bitwised modules")
	for i := 0; i < 3; i++ {
		log := <-anyWriter.LogChannel
		fmt.Printf("%v", log)
	}

	logPkg := logger.Init()
	logPkg.ConfigurePkg(anyWriter, "main")

	go generatePkgLogs(logPkg)
	fmt.Println(strings.Repeat("=", 5), "by package names")
	for i := 0; i < 3; i++ {
		log := <-anyWriter.LogChannel
		fmt.Printf("%v", log)
	}
}
