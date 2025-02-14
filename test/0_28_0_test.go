package test

import (
	"github.com/donnie4w/go-logger/logger"
	"testing"
)

func TestOption4mixed(t *testing.T) {
	logger.SetOption(&logger.Option{Console: true, FileOption: &logger.FileMixedMode{Filename: "testmixid.log", Maxbackup: 10, IsCompress: true, Timemode: logger.MODE_DAY, Maxsize: 1 << 20}})
	for i := 0; i < 10000; i++ {
		logger.Debug("this is a debug message", 1111111111111111111)
		logger.Info("this is a info message", 2222222222222222222)
		logger.Warn("this is a warn message", 33333333333333333)
		logger.Error("this is a error message", 4444444444444444444)
		logger.Fatal("this is a fatal message", 555555555555555555)
	}
}
