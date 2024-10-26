package test

import (
	"github.com/donnie4w/go-logger/logger"
	"testing"
	"time"
)

/*控制台打印，直接调用打印方法Debug(),Info()等方法*/
func Test_Global(t *testing.T) {
	logger.SetRollingDaily(``, "logger.log")
	// SetConsole(false)
	logger.SetFormat(logger.FORMAT_DATE | logger.FORMAT_LONGFILENAME) //设置后，下面日志格式只打印日期+短文件信息
	logger.Debug("this is debug message")
	logger.SetFormat(logger.FORMAT_DATE | logger.FORMAT_RELATIVEFILENAME) //设置后，下面日志格式只打印日期+短文件信息
	logger.Info("this is info message")
	logger.SetFormat(logger.FORMAT_DATE | logger.FORMAT_TIME | logger.FORMAT_SHORTFILENAME) //设置后，下面日志格式只打印日期+短文件信息
	logger.Warn("this is warning message")
	// SetLevel(FATAL) //设置为FATAL后，下面Error()级别小于FATAL,将不打印出来
	logger.Error("this is error message")
	logger.SetFormat(logger.FORMAT_LEVELFLAG | logger.FORMAT_DATE | logger.FORMAT_MICROSECONDS | logger.FORMAT_SHORTFILENAME)
	logger.SetFormatter("{message}|{level} {time} {file}\n")
	// SetFormat(FORMAT_NANO)
	logger.Fatal("this is fatal message")
}

/*设置日志文件*/
func Test_NewLogger(t *testing.T) {
	/*获取全局log单例，单日志文件项目日志建议使用单例*/
	//log := GetStaticLogger()
	/*获取新的log实例，要求不同日志文件时，使用多实例对象*/
	log := logger.NewLogger()
	/*按日期分割日志文件，也是默认设置值*/
	// log.SetRollingDaily(`D:\cfoldTest`, "log.txt")
	log.SetRollingByTime("", "newlogger.txt", logger.MODE_DAY)
	/*按日志文件大小分割日志文件*/
	// log.SetRollingFile("", "log1.txt", 3, KB)
	// log.SetRollingFileLoop(`D:\cfoldTest`, "log1.txt", 3, KB, 5)
	/* 设置打印级别 OFF,DEBUG,INFO,WARN,ERROR,FATAL
	log.SetLevel(OFF) 设置OFF后，将不再打印后面的日志 默认日志级别为ALL，打印级别*/
	/* 日志写入文件时，同时在控制台打印出来，设置为false后将不打印在控制台，默认值true*/
	// log.SetConsole(false)
	log.Debug("this is debug message")
	log.SetFormat(logger.FORMAT_LONGFILENAME) //设置后将打印出文件全部路径信息
	log.Info("this is info message")
	log.SetFormat(logger.FORMAT_MICROSECONDS | logger.FORMAT_SHORTFILENAME) //设置日志格式，时间+短文件名
	log.Warn("this is warning message")
	log.SetLevel(logger.LEVEL_FATAL) //设置为FATAL后，下面Error()级别小于FATAL,将不打印出来
	log.Error("this is error message")
	log.Fatal("this is fatal message")
	time.Sleep(2 * time.Second)
}

func Test_formatter(t *testing.T) {
	log := logger.NewLogger()
	log.SetFormatter("{time} {file} {level} >> {message}\n")
	log.Debug("this is debug message")
	log.Warn("this is info message")
	log.Error("this is error message")
	log.Fatal("this is fatal message")

	log.Debugf("this is debug message: %s", time.Now().Format("2006-01-02 15:04:05"))
}
