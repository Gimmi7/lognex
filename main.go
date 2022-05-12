package main

import (
	"github.com/Gimmi7/lognex/log"
	"go.uber.org/zap"
)

type Person struct {
	Name string
	Age  int
}

func main() {
	log.RecommendLoggerWithLogPath("d:/trashCan/logs")
	for i := 0; i < 10; i++ {
		log.Info("log test", zap.Int("index", i))
		log.Sugar().Debugf("%#v", &Person{Name: "jerry", Age: 11})
		log.Sugar().Errorf("server timeout, index=%d", i)
	}
}
