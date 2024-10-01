package logger

import (
	"testing"
)

func BenchmarkSerialLogger(b *testing.B) {
	log := NewLogger()
	log.SetRollingFile("", "logger.txt", 100, MB)
	log.SetConsole(false)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		log.Debug(i, ">>>this is debug message")
	}
}

func BenchmarkParallelLogger(b *testing.B) {
	log := NewLogger()
	log.SetRollingFile("", "logger.txt", 100, MB)
	log.SetConsole(false)
	b.ResetTimer()
	var i int64 = 0
	b.RunParallel(func(pb *testing.PB) {
		i++
		for pb.Next() {
			log.Debug(i, ">>>this is debug message")
		}
	})
}
