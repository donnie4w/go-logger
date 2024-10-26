package test

import (
	"github.com/donnie4w/go-logger/logger"
	"log/slog"
	"path/filepath"
	"strconv"
	"testing"
)

func TestSlog(t *testing.T) {
	replace := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			source.File = filepath.Base(source.File)
		}
		return a
	}
	loggingFile := logger.NewLogger()
	loggingFile.SetRollingFile("./1", "slogfile.txt", 100, logger.KB)
	h := slog.NewJSONHandler(loggingFile, &slog.HandlerOptions{AddSource: true, ReplaceAttr: replace})
	log := slog.New(h)
	for i := 0; i < 1000; i++ {
		log.Info("this is a info message:" + strconv.Itoa(i))
	}
}
