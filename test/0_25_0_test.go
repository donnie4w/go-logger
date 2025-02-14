package test

import (
	"github.com/donnie4w/go-logger/logger"
	"testing"
	"time"
)

func TestOption4time(t *testing.T) {
	logger.SetOption(&logger.Option{Level: logger.LEVEL_INFO, Console: true, FileOption: &logger.FileTimeMode{Filename: "testlogtime.log", Maxbackup: 3, IsCompress: false, Timemode: logger.MODE_MONTH}})
	for i := 0; i < 100; i++ {
		logger.Debug("this is a debug message", 1111111111111111111)
		logger.Info("this is a info message", 2222222222222222222)
		logger.Warn("this is a warn message", 33333333333333333)
		logger.Error("this is a error message", 4444444444444444444)
		logger.Fatal("this is a fatal message", 555555555555555555)
		time.Sleep(2 * time.Second)
	}
}

func TestOption4size(t *testing.T) {
	logger.SetOption(&logger.Option{Level: logger.LEVEL_DEBUG, Console: true, FileOption: &logger.FileSizeMode{Filename: "testlog.log", Maxsize: 1 << 10, Maxbackup: 10, IsCompress: false}})
	for i := 0; i < 20; i++ {
		logger.Debug("this is a debug message", 1111111111111111111)
		logger.Info("this is a info message", 2222222222222222222)
		logger.Warn("this is a warn message", 33333333333333333)
		logger.Error("this is a error message", 4444444444444444444)
		logger.Fatal("this is a fatal message", 555555555555555555)
	}
	time.Sleep(1 * time.Second)
}
