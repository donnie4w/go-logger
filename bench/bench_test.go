package bench

import (
	"github.com/donnie4w/go-logger/logger"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkSerialLogger(b *testing.B) {
	log := logger.NewLogger()
	log.SetRollingFile("", "logger2.log", 500, logger.MB)
	log.SetConsole(false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Debug(">>>>>>this is debug message>>>>>>this is debug message")
	}
}

func BenchmarkParallelLogger(b *testing.B) {
	log := logger.NewLogger()
	log.SetRollingFile("", "logger2.log", 200, logger.MB)
	log.SetConsole(false)
	b.SetParallelism(20)
	b.ResetTimer()
	var i int64 = 0
	b.RunParallel(func(pb *testing.PB) {
		if i == 30000 {
			return
		}
		i++
		for pb.Next() {
			log.Debug(">>>>>>this is debug message>>>>>>this is debug message")
		}

	})
}

// slog
func BenchmarkSerialSlog(b *testing.B) {
	replace := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			source.File = filepath.Base(source.File)
		}
		return a
	}
	out, _ := os.OpenFile(`slog.log`, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	h := slog.NewTextHandler(out, &slog.HandlerOptions{AddSource: true, ReplaceAttr: replace})
	log := slog.New(h)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info(">>>this is debug message")
	}
}

func BenchmarkParallelSLog(b *testing.B) {
	replace := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			source.File = filepath.Base(source.File)
		}
		return a
	}
	out, _ := os.OpenFile(`slog2.log`, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	h := slog.NewTextHandler(out, &slog.HandlerOptions{AddSource: true, ReplaceAttr: replace})
	log := slog.New(h)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Info(">>>this is debug message")
		}
	})
}

func BenchmarkMixedMode(b *testing.B) {
	goLogger := logger.NewLogger()
	goLogger.SetOption(&logger.Option{Level: logger.LEVEL_DEBUG, Console: false, FileOption: &logger.FileMixedMode{Filename: "testmixed.log", Maxsize: 200 << 20, Maxbackup: 10, IsCompress: false}})
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			goLogger.Info("this is info message")
		}
	})
}
