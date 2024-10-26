package test

import (
	"fmt"
	"github.com/donnie4w/go-logger/logger"
	"testing"
)

func TestCustomHandler(t *testing.T) {
	levelname := func(level logger.LEVELTYPE) string {
		switch level {
		case logger.LEVEL_DEBUG:
			return "debug"
		case logger.LEVEL_INFO:
			return "info"
		case logger.LEVEL_FATAL:
			return "fatal"
		case logger.LEVEL_WARN:
			return "warn"
		case logger.LEVEL_ERROR:
			return "error"
		default:
			return "unknown"
		}
	}

	logger.SetOption(&logger.Option{Console: true, CustomHandler: func(lc *logger.LogContext) bool {
		fmt.Println("level:", levelname(lc.Level))
		fmt.Println("message:", fmt.Sprint(lc.Args...))
		if lc.Level == logger.LEVEL_ERROR {
			return false //if error mesaage , do not print
		}
		return true
	},
	})
	logger.Debug("this is a debug message")
	logger.Info("this is a info message")
	logger.Warn("this is a warn message")
	logger.Error("this is a error message")
}

func TestStacktrace(t *testing.T) {
	logger.SetOption(&logger.Option{Console: true, Stacktrace: logger.LEVEL_WARN, Format: logger.FORMAT_LEVELFLAG | logger.FORMAT_DATE | logger.FORMAT_TIME | logger.FORMAT_SHORTFILENAME | logger.FORMAT_FUNC})
	logger.Debug("this is a debug message")
	Stacktrace1()
}

func Stacktrace1() {
	logger.Info("this is a info message")
	Stacktrace2()
}

func Stacktrace2() {
	logger.Warn("this is a warn message")
	Stacktrace3()
}

func Stacktrace3() {
	logger.Error("this is a error message")
	logger.Fatal("this is a fatal message")
}

func TestLevelOptions(t *testing.T) {
	logger.SetLevelOption(logger.LEVEL_DEBUG, &logger.LevelOption{Format: logger.FORMAT_LEVELFLAG | logger.FORMAT_TIME | logger.FORMAT_SHORTFILENAME})
	logger.SetLevelOption(logger.LEVEL_INFO, &logger.LevelOption{Format: logger.FORMAT_LEVELFLAG})
	logger.SetLevelOption(logger.LEVEL_WARN, &logger.LevelOption{Format: logger.FORMAT_LEVELFLAG | logger.FORMAT_TIME | logger.FORMAT_SHORTFILENAME | logger.FORMAT_DATE | logger.FORMAT_FUNC})

	logger.Debug("this is a debug message")
	logger.Info("this is a info message")
	logger.Warn("this is a warn message")
}
