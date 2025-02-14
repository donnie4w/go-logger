package test

import (
	"github.com/donnie4w/go-logger/logger"
	"testing"
	"time"
)

func Test_AttrFormat(t *testing.T) {
	attrformat := &logger.AttrFormat{
		SetLevelFmt: func(level logger.LEVELTYPE) string {
			switch level {
			case logger.LEVEL_DEBUG:
				return "debug:"
			case logger.LEVEL_INFO:
				return "info:"
			case logger.LEVEL_WARN:
				return "warn:"
			case logger.LEVEL_ERROR:
				return "error>>>>"
			case logger.LEVEL_FATAL:
				return "[fatal]"
			default:
				return "[unknown]"
			}
		},
		SetTimeFmt: func() (string, string, string) {
			s := time.Now().Format("2006-01-02 15:04:05")
			return s, "", ""
		},
	}
	logger.SetOption(&logger.Option{AttrFormat: attrformat, Console: true, FileOption: &logger.FileTimeMode{Filename: "testlogtime.log", Maxbackup: 3, IsCompress: false, Timemode: logger.MODE_MONTH}})
	logger.Debug("this is a debug message", 1111111111111111111)
	logger.Info("this is a info message", 2222222222222222222)
	logger.Warn("this is a warn message", 33333333333333333)
	logger.Error("this is a error message", 4444444444444444444)
	logger.Fatal("this is a fatal message", 555555555555555555)
}

func Test_AttrFormat2(t *testing.T) {
	attrformat := &logger.AttrFormat{
		SetBodyFmt: func(level logger.LEVELTYPE, bs []byte) []byte {
			//处理日志末尾换行符
			if size := len(bs); bs[size-1] == '\n' {
				bs = append(bs[:size-1], []byte("\x1b[0m\n")...)
			} else {
				bs = append(bs, []byte("\x1b[0m\n")...)
			}
			switch level {
			case logger.LEVEL_DEBUG:
				return append([]byte("\x1b[34m"), bs...)
			case logger.LEVEL_INFO:
				return append([]byte("\x1b[32m"), bs...)
			case logger.LEVEL_WARN:
				return append([]byte("\x1b[33m"), bs...)
			case logger.LEVEL_ERROR:
				return append([]byte("\x1b[31m"), bs...)
			case logger.LEVEL_FATAL:
				return append([]byte("\x1b[41m"), bs...)
			default:
				return bs
			}
		},
	}
	logger.SetOption(&logger.Option{AttrFormat: attrformat, Console: true, FileOption: &logger.FileTimeMode{Filename: "testlogtime.log", Maxbackup: 3, IsCompress: false, Timemode: logger.MODE_MONTH}})
	logger.Debug("this is a debug message:", 111111111111111110)
	logger.Info("this is a info message:", 222222222222222220)
	logger.Warn("this is a warn message:", 333333333333333330)
	logger.Error("this is a error message:", 4444444444444444440)
	logger.Fatal("this is a fatal message:", 5555555555555555550)
}

func Test_format(t *testing.T) {
	logger.SetRollingFile("", "test.log", 10, logger.MB)
	logger.Debugf("this is a debugf message:%d", 1)
	logger.Infof("this is a infof message:%s", "hi,logger")
	logger.Warnf("this is a warnf message:%x,%x", 14, 15)
	logger.Errorf("this is a errorf message:%f", 44.4444)
	logger.Fatalf("this is a fatalf message:%t", true)
	logger.Debugf("this is a debugf message:%p", new(int))
}
