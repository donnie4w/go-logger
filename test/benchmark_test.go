package test

import (
	"github.com/donnie4w/go-logger/logger"
	"testing"
)

var (
	goLogger = logger.NewLogger()
)

func init() {
	goLogger.SetConsole(false)
	goLogger.SetRollingFile("log", "test.log", 10, logger.MB)
	goLogger.SetGzipOn(true)
}
func BenchmarkRolling(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// do something
			goLogger.Info("this is info message")
		}
	})
}
