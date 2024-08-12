package logger

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
)

/*控制台打印，直接调用打印方法Debug(),Info()等方法*/
func Test_Log(t *testing.T) {
	SetRollingDaily(`D:\cfoldTest`, "log2.txt")
	// SetConsole(false)
	Debug("this is debug message")
	Info("this is info message")
	SetFormat(FORMAT_DATE | FORMAT_SHORTFILENAME) //设置后，下面日志格式只打印日期+短文件信息
	Warn("this is warning message")
	// SetLevel(FATAL) //设置为FATAL后，下面Error()级别小于FATAL,将不打印出来
	Error("this is error message")
	SetFormat(FORMAT_LEVELFLAG | FORMAT_DATE | FORMAT_MICROSECONDS | FORMAT_SHORTFILENAME)
	SetFormatter("{message}|{level} {time} {file}\n")
	// SetFormat(FORMAT_NANO)
	Fatal("this is fatal message")
}

/*设置日志文件*/
func Test_LogOne(t *testing.T) {
	/*获取全局log单例，单日志文件项目日志建议使用单例*/
	//log := GetStaticLogger()
	/*获取新的log实例，要求不同日志文件时，使用多实例对象*/
	log := NewLogger()
	/*按日期分割日志文件，也是默认设置值*/
	// log.SetRollingDaily(`D:\cfoldTest`, "log.txt")
	log.SetRollingByTime(`D:\cfoldTest`, "log.txt", MODE_DAY)
	/*按日志文件大小分割日志文件*/
	// log.SetRollingFile("", "log1.txt", 3, KB)
	// log.SetRollingFileLoop(`D:\cfoldTest`, "log1.txt", 3, KB, 5)
	/* 设置打印级别 OFF,DEBUG,INFO,WARN,ERROR,FATAL
	log.SetLevel(OFF) 设置OFF后，将不再打印后面的日志 默认日志级别为ALL，打印级别*/
	/* 日志写入文件时，同时在控制台打印出来，设置为false后将不打印在控制台，默认值true*/
	// log.SetConsole(false)
	log.Debug("this is debug message")
	log.SetFormat(FORMAT_LONGFILENAME) //设置后将打印出文件全部路径信息
	log.Info("this is info message")
	log.SetFormat(FORMAT_MICROSECONDS | FORMAT_SHORTFILENAME) //设置日志格式，时间+短文件名
	log.Warn("this is warning message")
	log.SetLevel(LEVEL_FATAL) //设置为FATAL后，下面Error()级别小于FATAL,将不打印出来
	log.Error("this is error message")
	log.Fatal("this is fatal message")
	time.Sleep(2 * time.Second)
}

func BenchmarkSerialLog(b *testing.B) {
	b.StopTimer()
	log := NewLogger()
	log.SetRollingFile(`./`, "log1.txt", 100, MB)
	log.SetConsole(false)
	// log.SetFormat(FORMAT_NANO)
	b.StartTimer()
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			for i := 0; i < b.N; i++ {
				// log.Write([]byte(">>>aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
				log.Debug(i, ">>>this is debug message")
				// log.Info(i, ">>>this is info message")
				// log.Warn(i, ">>>this is warm message")
				// log.log.Error(i, ">>>this is error message")
			}
			wg.Done()
		}()
	}
	wg.Wait()

}

func TestSlog(t *testing.T) {
	replace := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			source.File = filepath.Base(source.File)
		}
		return a
	}
	loggingFile := NewLogger()
	loggingFile.SetRollingFile("./1", "slogfile.txt", 100, KB)
	h := slog.NewJSONHandler(loggingFile, &slog.HandlerOptions{AddSource: true, ReplaceAttr: replace})
	log := slog.New(h)
	for i := 0; i < 1000; i++ {
		log.Info("this is a info message:" + strconv.Itoa(i))
	}
}

func TestOption4time(t *testing.T) {
	SetOption(&Option{Level: LEVEL_INFO, Console: true, FileOption: &FileTimeMode{Filename: "testlogtime.log", Maxbuckup: 3, IsCompress: false, Timemode: MODE_MONTH}})
	for i := 0; i < 100; i++ {
		Debug("this is a debug message", 1111111111111111111)
		Info("this is a info message", 2222222222222222222)
		time.Sleep(2 * time.Second)
	}
}

func TestOption4size(t *testing.T) {
	SetOption(&Option{Level: LEVEL_DEBUG, Console: true, FileOption: &FileSizeMode{Filename: "testlog.log", Maxsize: 500, Maxbuckup: 3, IsCompress: false}})
	for i := 0; i < 20; i++ {
		Debug("this is a debug message", 1111111111111111111)
		time.Sleep(100 * time.Millisecond)
	}
}

func TestCustomHandler(t *testing.T) {
	SetOption(&Option{Console: true, CustomHandler: func(lc *LogContext) bool {
		fmt.Println("level:", levelname(lc.Level))
		fmt.Println("message:", fmt.Sprint(lc.Args...))
		if lc.Level == LEVEL_ERROR {
			return false //if error mesaage , do not print
		}
		return true
	},
	})
	Debug("this is a debug message")
	Info("this is a info message")
	Warn("this is a warn message")
	Error("this is a error message")
}

func levelname(level _LEVEL) string {
	switch level {
	case LEVEL_DEBUG:
		return "debug"
	case LEVEL_INFO:
		return "info"
	case LEVEL_FATAL:
		return "fatal"
	case LEVEL_WARN:
		return "warn"
	case LEVEL_ERROR:
		return "error"
	default:
		return "unknown"
	}
}

func TestStacktrace(t *testing.T) {
	SetOption(&Option{Console: true, Stacktrace: LEVEL_WARN, Format: FORMAT_LEVELFLAG | FORMAT_DATE | FORMAT_TIME | FORMAT_SHORTFILENAME | FORMAT_FUNC})
	Debug("this is a debug message")
	Stacktrace1()
}

func Stacktrace1() {
	Info("this is a info message")
	Stacktrace2()
}

func Stacktrace2() {
	Warn("this is a warn message")
	Stacktrace3()
}

func Stacktrace3() {
	Error("this is a error message")
	Fatal("this is a fatal message")
}

func TestLevelOptions(t *testing.T) {
	SetLevelOption(LEVEL_DEBUG, &LevelOption{Format: FORMAT_LEVELFLAG | FORMAT_TIME | FORMAT_SHORTFILENAME})
	SetLevelOption(LEVEL_INFO, &LevelOption{Format: FORMAT_LEVELFLAG})
	SetLevelOption(LEVEL_WARN, &LevelOption{Format: FORMAT_LEVELFLAG | FORMAT_TIME | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_FUNC})

	Debug("this is a debug message")
	Info("this is a info message")
	Warn("this is a warn message")
}
