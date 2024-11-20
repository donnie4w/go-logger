package bench

import (
	"github.com/donnie4w/go-logger/logger"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"sync"
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
	log.SetRollingFile("", "logger2.log", 50, logger.MB)
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

func Test_Bench(t *testing.T) {
	go http.ListenAndServe(":9000", nil)
	log := logger.NewLogger()
	log.SetRollingFile("", "loggerBench.log", 100, logger.MB)
	log.SetConsole(false)
	for i := 0; i < 1<<30; i++ {
		var wg sync.WaitGroup
		for j := 0; j < 1000; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				log.Debug(i, ">>>this is debug message", "!!!11111111111111111111111111")
			}()
		}
		wg.Wait()
	}
}

func Test_Bench2(t *testing.T) {
	log := logger.NewLogger()
	log.SetRollingFile("", "logger2.log", 500, logger.MB)
	log.SetConsole(false)
	for i := 0; i < 10; i++ {
		log.Debug(">>>>>>this is debug message>>>>>>this is debug message")
	}
}

var (
	goLogger = logger.NewLogger()
)

func init() {
	goLogger.SetOption(&logger.Option{Level: logger.LEVEL_DEBUG, Console: true, FileOption: &logger.FileMixedMode{Filename: "test.log", Maxsize: 20, SizeUint: logger.MB, Maxbuckup: 1, IsCompress: true}})
	goLogger.SetConsole(false)
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
