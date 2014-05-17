package logtest

import (
	"github.com/tomcatzh/go-logger/logger"
	"runtime"
	"strconv"
	"testing"
)

func simpleLog(i int, log bool) int {
	i += 1
	if log {
		logger.Debug("Debug>>>>>>>>>>>>>>>>>>>>>>" + strconv.Itoa(i))
		logger.Info("Info>>>>>>>>>>>>>>>>>>>>>>>>>" + strconv.Itoa(i))
		logger.Warn("Warn>>>>>>>>>>>>>>>>>>>>>>>>>" + strconv.Itoa(i))
		logger.Error("Error>>>>>>>>>>>>>>>>>>>>>>>>>" + strconv.Itoa(i))
		logger.Fatal("Fatal>>>>>>>>>>>>>>>>>>>>>>>>>" + strconv.Itoa(i))
	}
	return i
}

func Test_SimpleLog(t *testing.T) {
	logger.SetConsole(true)
	logger.SetRollingDaily("test", "test.log", 0644)
	logger.SetLevel(logger.ERROR)
	r := simpleLog(1, true)
	if r != 2 {
		t.Error("simpleLog error!")
	}
}

const workerCount = 20

type work struct {
	in  int
	out int
}

func logWorker(in <-chan work, out chan<- work, quit chan bool, log bool) {
	for {
		select {
		case s := <-in:
			s.out = simpleLog(s.in, log)
			out <- s
		case <-quit:
			return
		}
	}
}

func logSender(n int, in chan<- work) {
	for i := 0; i < n; i++ {
		in <- work{i, 0}
	}
}

func doMultiTest(b *testing.B, log bool) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	in := make(chan work, 100)
	out := make(chan work, 100)
	quit := make(chan bool, workerCount)
	for i := 0; i < workerCount; i++ {
		go logWorker(in, out, quit, log)
	}

	go logSender(b.N, in)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := <-out
		if w.out != w.in+1 {
			b.Error("simpleLog error!")
		}
	}
	b.StopTimer()

	for i := 0; i < workerCount; i++ {
		quit <- true
	}
	runtime.GOMAXPROCS(1)
}

func Benchmark_Multi_NoLog(b *testing.B) {
	doMultiTest(b, false)
}

type fileMode int

const (
	dailyFile   fileMode = 1
	rollingFile fileMode = 2
)

func setFile(fileType fileMode, filename string) {
	dir := "test"

	maxNumber := int32(3)
	maxSize := int64(10)
	unit := logger.MB

	switch fileType {
	case dailyFile:
		logger.SetRollingDaily(dir, filename, 0644)
	case rollingFile:
		logger.SetRollingFile(dir, filename, maxNumber, maxSize, unit, 0644)
	}
}

func Benchmark_Multi_Console_Daily_Off(b *testing.B) {
	logger.SetConsole(true)
	setFile(dailyFile, "multi_console_daily_off.log")
	logger.SetLevel(logger.OFF)
	b.ResetTimer()
	doMultiTest(b, true)
}

func Benchmark_Multi_NoConsole_Daily_Off(b *testing.B) {
	logger.SetConsole(false)
	setFile(dailyFile, "multi_noconsole_daily_off.log")
	logger.SetLevel(logger.OFF)
	b.ResetTimer()
	doMultiTest(b, true)
}

func Benchmark_Multi_Console_Daily_All(b *testing.B) {
	logger.SetConsole(true)
	setFile(dailyFile, "multi_console_daily_all.log")
	logger.SetLevel(logger.ALL)
	b.ResetTimer()
	doMultiTest(b, true)
}

func Benchmark_Multi_NoConsole_Daily_All(b *testing.B) {
	logger.SetConsole(false)
	setFile(dailyFile, "multi_noconsole_daily_all.log")
	logger.SetLevel(logger.ALL)
	b.ResetTimer()
	doMultiTest(b, true)
}

func Benchmark_Multi_Console_File_All(b *testing.B) {
	logger.SetConsole(true)
	setFile(rollingFile, "multi_console_file_all.log")
	logger.SetLevel(logger.ALL)
	b.ResetTimer()
	doMultiTest(b, true)
}

func Benchmark_Multi_NoConsole_File_All(b *testing.B) {
	logger.SetConsole(false)
	setFile(rollingFile, "multi_noconsole_file_all.log")
	logger.SetLevel(logger.ALL)
	b.ResetTimer()
	doMultiTest(b, true)
}

func Benchmark_Multi_Console_Daily_Error(b *testing.B) {
	logger.SetConsole(true)
	setFile(dailyFile, "multi_console_daily_error.log")
	logger.SetLevel(logger.ERROR)
	b.ResetTimer()
	doMultiTest(b, true)
}

func Benchmark_Multi_NoConsole_Daily_Error(b *testing.B) {
	logger.SetConsole(false)
	setFile(dailyFile, "multi_noconsole_daily_error.log")
	logger.SetLevel(logger.ERROR)
	b.ResetTimer()
	doMultiTest(b, true)
}

func doSingleTest(b *testing.B, log bool) {
	for i := 0; i < b.N; i++ {
		r := simpleLog(i, log)
		if r != i+1 {
			b.Error("simpleLog error!")
		}
	}
}

func Benchmark_Single_NoLog(b *testing.B) {
	doSingleTest(b, false)
}

func Benchmark_Single_Console_Daily_Off(b *testing.B) {
	logger.SetConsole(true)
	setFile(dailyFile, "single_console_daily_off.log")
	logger.SetLevel(logger.OFF)
	b.ResetTimer()
	doSingleTest(b, true)
}

func Benchmark_Single_NoConsole_Daily_Off(b *testing.B) {
	logger.SetConsole(false)
	setFile(dailyFile, "single_noconsole_daily_off.log")
	logger.SetLevel(logger.OFF)
	b.ResetTimer()
	doSingleTest(b, true)
}

func Benchmark_Single_Console_Daily_All(b *testing.B) {
	logger.SetConsole(true)
	setFile(dailyFile, "single_console_daily_all.log")
	logger.SetLevel(logger.ALL)
	b.ResetTimer()
	doSingleTest(b, true)
}

func Benchmark_Single_NoConsole_Daily_All(b *testing.B) {
	logger.SetConsole(false)
	setFile(dailyFile, "single_noconsole_daily_all.log")
	logger.SetLevel(logger.ALL)
	b.ResetTimer()
	doSingleTest(b, true)
}

func Benchmark_Single_Console_File_All(b *testing.B) {
	logger.SetConsole(true)
	setFile(rollingFile, "single_console_file_all.log")
	logger.SetLevel(logger.ALL)
	b.ResetTimer()
	doSingleTest(b, true)
}

func Benchmark_Single_NoConsole_File_All(b *testing.B) {
	logger.SetConsole(false)
	setFile(rollingFile, "single_noconsole_file_all.log")
	logger.SetLevel(logger.ALL)
	b.ResetTimer()
	doSingleTest(b, true)
}

func Benchmark_Single_Console_Daily_Error(b *testing.B) {
	logger.SetConsole(true)
	setFile(dailyFile, "single_console_daily_error.log")
	logger.SetLevel(logger.ERROR)
	b.ResetTimer()
	doSingleTest(b, true)
}

func Benchmark_Single_NoConsole_Daily_Error(b *testing.B) {
	logger.SetConsole(false)
	setFile(dailyFile, "single_noconsole_daily_error.log")
	logger.SetLevel(logger.ERROR)
	b.ResetTimer()
	doSingleTest(b, true)
}
