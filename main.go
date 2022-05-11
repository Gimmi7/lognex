package main

import (
	"go.uber.org/zap"
	"lognex/log"
)

func main() {
	log.RecommendLoggerWithLogPath("d:/trashCan/logs")
	for i := 0; i < 10; i++ {
		log.Info("log test", zap.Int("index", i))
		log.Sugar().Errorf("server timeout, index=%d", i)
	}
}
