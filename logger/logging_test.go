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
	Debug("11111111111111")
	Info("22222222")
	SetFormat(FORMAT_DATE | FORMAT_SHORTFILENAME) //设置后，下面日志格式只打印日期+短文件信息
	Warn("333333333")
	// SetLevel(FATAL) //设置为FATAL后，下面Error()级别小于FATAL,将不打印出来
	Error("444444444")
	SetFormat(FORMAT_LEVELFLAG | FORMAT_DATE | FORMAT_MICROSECNDS | FORMAT_SHORTFILENAME)
	SetFormatter("{message}|{level} {time} {file}\n")
	// SetFormat(FORMAT_NANO)
	Fatal("5555555555")
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
	log.Debug("aaaaaaaaaaaaaaaaaaaaaaaa")
	log.SetFormat(FORMAT_LONGFILENAME) //设置后将打印出文件全部路径信息
	log.Info("bbbbbbbbbbbbbbbbbbbbbbbb")
	log.SetFormat(FORMAT_MICROSECNDS | FORMAT_SHORTFILENAME) //设置日志格式，时间+短文件名
	log.Warn("ccccccccccccccccccccccc")
	log.SetLevel(LEVEL_FATAL) //设置为FATAL后，下面Error()级别小于FATAL,将不打印出来
	log.Error("ddddddddddddddddddddddd")
	log.Fatal("eeeeeeeeeeeeeeeeeeeeeee")
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
				log.Debug(i, ">>>aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
				// log.Info(i, ">>>bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
				// log.Warn(i, ">>>cccccccccccccccccccccccccccccccccccc")
				// log.log.Error(i, ">>>dddddddddddddddddddddddddddddddddddd")
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
		log.Info(">>>aaaaaaaaaaaaaaaaaaaaaaa:" + strconv.Itoa(i))
	}
}

func TestOption4time(t *testing.T) {
	SetOption(&Option{Level: LEVEL_INFO, Console: true, FileOption: &FileTimeMode{Filename: "testlogtime.log", Maxbuckup: 3, IsCompress: false, Timemode: MODE_MONTH}})
	for i := 0; i < 100; i++ {
		Debug("aaaaaaaaaa", 1111111111111111111)
		Info("bbbbbbbbbb", 2222222222222222222)
		time.Sleep(2 * time.Second)
	}
}

func TestOption4size(t *testing.T) {
	SetOption(&Option{Level: LEVEL_DEBUG, Console: true, FileOption: &FileSizeMode{Filename: "testlog.log", Maxsize: 500, Maxbuckup: 3, IsCompress: false}})
	for i := 0; i < 20; i++ {
		Debug("bbbbbbbbbb", 222222222)
		time.Sleep(100 * time.Millisecond)
	}
}

func TestOptionHandler(t *testing.T) {
	SetOption(&Option{Level: LEVEL_DEBUG, Console: true,
		FileOption: &FileSizeMode{Filename: "testlog.log", Maxsize: 500, Maxbuckup: 3, IsCompress: false},
		CustomHandler: func(lc *LogContext) bool {
			fmt.Println(1)
			return true
			//return false
		},
	})
	for i := 0; i < 20; i++ {
		Debug("bbbbbbbbbb", 222222222)
		time.Sleep(100 * time.Millisecond)
	}
}
