// Copyright (c) 2014, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/go-logger

package logger

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/donnie4w/gofer/buffer"
	. "github.com/donnie4w/gofer/fastio"
	"github.com/donnie4w/gofer/hashmap"
)

type LEVELTYPE int8
type _UNIT int64
type _MODE_TIME uint8
type _CUTMODE int //dailyRolling ,rollingFile
type _FORMAT int

const (
	_DATEFORMAT_DAY   = "20060102"
	_DATEFORMAT_HOUR  = "2006010215"
	_DATEFORMAT_MONTH = "200601"
	default_format    = FORMAT_LEVELFLAG | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME
	default_level     = LEVEL_ALL
)

var static_lo = NewLogger()

var TIME_DEVIATION time.Duration

const (
	_        = iota
	KB _UNIT = 1 << (iota * 10)
	MB
	GB
	TB
)

const (
	MODE_HOUR  _MODE_TIME = 1
	MODE_DAY   _MODE_TIME = 2
	MODE_MONTH _MODE_TIME = 3
)

const (
	// FORMAT_NANO
	//
	// no format, Only log content is printed
	// 无其他格式，只打印日志内容
	FORMAT_NANO _FORMAT = 64

	// FORMAT_LONGFILENAME
	//
	// full file name and line number
	// 长文件名(文件绝对路径)及行数
	FORMAT_LONGFILENAME = _FORMAT(8)

	// FORMAT_SHORTFILENAME
	//
	// final file name element and line number
	// 短文件名及行数
	FORMAT_SHORTFILENAME = _FORMAT(16)

	// FORMAT_RELATIVEFILENAME
	//
	// relative file name element and line number
	// 相对路径文件名及行数
	FORMAT_RELATIVEFILENAME = _FORMAT(256)

	// FORMAT_DATE
	//
	// the date in the local time zone: 2009/01/23
	// 日期时间精确到天
	FORMAT_DATE = _FORMAT(1)

	// FORMAT_TIME
	//
	// the time in the local time zone: 01:23:23
	// 时间精确到秒
	FORMAT_TIME = _FORMAT(2)

	// FORMAT_MICROSECONDS
	//
	// microsecond resolution: 01:23:23.123123.
	// 时间精确到微秒
	FORMAT_MICROSECONDS = _FORMAT(4)

	// FORMAT_LEVELFLAG
	//
	//Log level flag. e.g. [DEBUG],[INFO],[WARN],[ERROR],[FATAL]
	// 日志级别表示
	FORMAT_LEVELFLAG = _FORMAT(32)

	// FORMAT_FUNC
	//
	// the func of caller
	// 调用的函数名
	FORMAT_FUNC = _FORMAT(128)
)

const (

	// LEVEL_ALL is the lowest level,If the log level is this level, logs of other levels can be printed
	// 日志级别：ALL 打印所有日志
	LEVEL_ALL LEVELTYPE = iota

	// LEVEL_DEBUG  debug log level
	// 日志级别：DEBUG 小于INFO
	LEVEL_DEBUG

	// LEVEL_INFO info log level
	// 日志级别：INFO 小于 WARN
	LEVEL_INFO

	// LEVEL_WARN warn log level
	// 日志级别：WARN 小于 ERROR
	LEVEL_WARN

	// LEVEL_ERROR error log level
	// 日志级别：ERROR 小于 FATAL
	LEVEL_ERROR

	// LEVEL_FATAL fatal log level
	// 日志级别：FATAL 小于 OFF
	LEVEL_FATAL

	// LEVEL_OFF  means none of the logs can be printed
	// 日志级别：off 不打印任何日志
	LEVEL_OFF
)

var _DEBUG, _INFO, _WARN, _ERROR, _FATALE = []byte("[DEBUG]"), []byte("[INFO]"), []byte("[WARN]"), []byte("[ERROR]"), []byte("[FATAL]")

const (
	_TIMEMODE  _CUTMODE = 1
	_SIZEMODE  _CUTMODE = 2
	_MIXEDMODE _CUTMODE = 3
)

// 使用常量定义标志位组合
const (
	timeFlags = FORMAT_DATE | FORMAT_TIME | FORMAT_MICROSECONDS
	fileFlags = FORMAT_SHORTFILENAME | FORMAT_LONGFILENAME | FORMAT_RELATIVEFILENAME
)

// SetFormat sets the logging format to the specified format type.
//
// 设置打印格式: FORMAT_LEVELFLAG | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME
//
// Parameters:
//   - format: The desired format for log entries, represented by the _FORMAT type.
//
// Returns:
//   - *Logging: A Logging instance to enable method chaining.
func SetFormat(format _FORMAT) *Logging {
	return static_lo.SetFormat(format)
}

// SetLevel sets the logging level to the specified level type.
//
// 设置控制台日志级别，默认 LEVEL_ALL, 其他: LEVEL_DEBUG, LEVEL_INFO, LEVEL_WARN
//
// Parameters:
//   - level: The logging level (e.g., LEVEL_DEBUG, LEVEL_INFO, LEVEL_WARN), represented by the LEVELTYPE type.
//
// Returns:
//   - *Logging: A Logging instance to enable method chaining.
func SetLevel(level LEVELTYPE) *Logging {
	return static_lo.SetLevel(level)
}

// SetFormatter specifies a custom format string for the logging output.
//
// 设置输出格式，默认:  "{level}{time} {file} {message}\n"
//
// Parameters:
//   - formatter: A string defining the format for log entries, allowing custom log entry layouts.
//
// Returns:
//   - *Logging: A Logging instance to enable method chaining.
func SetFormatter(formatter string) *Logging {
	return static_lo.SetFormatter(formatter)
}

// SetConsole print logs on the console or not. default true
func SetConsole(on bool) *Logging {
	return static_lo.SetConsole(on)

}

// GetStaticLogger
// return the default log object
// 获得全局Logger对象
func GetStaticLogger() *Logging {
	return static_lo
}

// SetRollingFile when the log file(fileDir+`\`+fileName) exceeds the specified size(maxFileSize), it will be backed up with a specified file name
// Parameters:
//   - fileDir   :directory where log files are stored, If it is the current directory, you also can set it to ""
//   - fileName  : log file name
//   - maxFileSize :  maximum size of a log file
//   - unit		   :  size unit :  KB,MB,GB,TB
//
// Use SetOption() instead.
func SetRollingFile(fileDir, fileName string, maxFileSize int64, unit _UNIT) (l *Logging, err error) {
	return SetRollingFileLoop(fileDir, fileName, maxFileSize, unit, 0)
}

// SetRollingDaily yesterday's log data is backed up to a specified log file each day
// Parameters:
//   - fileDir   :directory where log files are stored, If it is the current directory, you also can set it to ""
//   - fileName  : log file name
//
// Use SetOption() instead.
func SetRollingDaily(fileDir, fileName string) (l *Logging, err error) {
	return SetRollingByTime(fileDir, fileName, MODE_DAY)
}

// SetRollingFileLoop like SetRollingFile,but only keep (maxFileNum) current files
// - maxFileNum : the number of files that are retained
//
// Use SetOption() instead.
func SetRollingFileLoop(fileDir, fileName string, maxFileSize int64, unit _UNIT, maxFileNum int) (l *Logging, err error) {
	return static_lo.SetRollingFileLoop(fileDir, fileName, maxFileSize, unit, maxFileNum)
}

// SetRollingByTime like SetRollingDaily,but supporte hourly backup ,dayly backup and monthly backup
// mode : 	MODE_HOUR    MODE_DAY   MODE_MONTH
//
// Use SetOption() instead.
func SetRollingByTime(fileDir, fileName string, mode _MODE_TIME) (l *Logging, err error) {
	return static_lo.SetRollingByTime(fileDir, fileName, mode)
}

// SetGzipOn when set true, the specified backup file of both SetRollingFile and SetRollingFileLoop will be save as a compressed file
//
// Use SetOption() instead.
func SetGzipOn(is bool) (l *Logging) {
	return static_lo.SetGzipOn(is)
}

// SetOption 配置对象
//
// e.g.
//
// SetOption(&Option{Level: LEVEL_DEBUG, Console: true, FileOption: &FileSizeMode{Filename: "test.log", Maxsize: 500, Maxbackup: 3, IsCompress: false}})
func SetOption(option *Option) *Logging {
	return static_lo.SetOption(option)
}

// Debug logs a message at the DEBUG level using the default logging instance.
// Accepts any number of arguments to format the log entry.
//
// Parameters:
//   - v: Variadic arguments to be logged.
//
// Returns:
//   - *Logging: A Logging instance for possible further usage.
func Debug(v ...any) *Logging {
	return println(nil, LEVEL_DEBUG, 2, v...)
}

// Info logs a message at the INFO level using the default logging instance.
// Accepts any number of arguments to format the log entry.
//
// Parameters:
//   - v: Variadic arguments to be logged.
//
// Returns:
//   - *Logging: A Logging instance for possible further usage.
func Info(v ...any) *Logging {
	return println(nil, LEVEL_INFO, 2, v...)
}

// Warn logs a message at the WARN level using the default logging instance.
// Accepts any number of arguments to format the log entry.
//
// Parameters:
//   - v: Variadic arguments to be logged.
//
// Returns:
//   - *Logging: A Logging instance for possible further usage.
func Warn(v ...any) *Logging {
	return println(nil, LEVEL_WARN, 2, v...)
}

// Error logs a message at the ERROR level using the default logging instance.
// Accepts any number of arguments to format the log entry.
//
// Parameters:
//   - v: Variadic arguments to be logged.
//
// Returns:
//   - *Logging: A Logging instance for possible further usage.
func Error(v ...any) *Logging {
	return println(nil, LEVEL_ERROR, 2, v...)
}

// Fatal logs a message at the FATAL level using the default logging instance and may terminate the application.
// Accepts any number of arguments to format the log entry.
//
// Parameters:
//   - v: Variadic arguments to be logged.
//
// Returns:
//   - *Logging: A Logging instance for possible further usage.
func Fatal(v ...any) *Logging {
	return println(nil, LEVEL_FATAL, 2, v...)
}

// Debugf logs a formatted message at the DEBUG level using the default logging instance.
// Takes a format string followed by variadic arguments to be formatted.
//
// Parameters:
//   - format: The format string for the log entry.
//   - v: Variadic arguments to be formatted according to the format string.
//
// Returns:
//   - *Logging: A Logging instance for possible further usage.
func Debugf(format string, v ...any) *Logging {
	return println(&format, LEVEL_DEBUG, 2, v...)
}

// Infof logs a formatted message at the INFO level using the default logging instance.
// Takes a format string followed by variadic arguments to be formatted.
//
// Parameters:
//   - format: The format string for the log entry.
//   - v: Variadic arguments to be formatted according to the format string.
//
// Returns:
//   - *Logging: A Logging instance for possible further usage.
func Infof(format string, v ...any) *Logging {
	return println(&format, LEVEL_INFO, 2, v...)
}

// Warnf logs a formatted message at the WARN level using the default logging instance.
// Takes a format string followed by variadic arguments to be formatted.
//
// Parameters:
//   - format: The format string for the log entry.
//   - v: Variadic arguments to be formatted according to the format string.
//
// Returns:
//   - *Logging: A Logging instance for possible further usage.
func Warnf(format string, v ...any) *Logging {
	return println(&format, LEVEL_WARN, 2, v...)
}

// Errorf logs a formatted message at the ERROR level using the default logging instance.
// Takes a format string followed by variadic arguments to be formatted.
//
// Parameters:
//   - format: The format string for the log entry.
//   - v: Variadic arguments to be formatted according to the format string.
//
// Returns:
//   - *Logging: A Logging instance for possible further usage.
func Errorf(format string, v ...any) *Logging {
	return println(&format, LEVEL_ERROR, 2, v...)
}

// Fatalf logs a formatted message at the FATAL level using the default logging instance and may terminate the application.
// Takes a format string followed by variadic arguments to be formatted.
//
// Parameters:
//   - format: The format string for the log entry.
//   - v: Variadic arguments to be formatted according to the format string.
//
// Returns:
//   - *Logging: A Logging instance for possible further usage.
func Fatalf(format string, v ...any) *Logging {
	return println(&format, LEVEL_FATAL, 2, v...)
}

func println(format *string, level LEVELTYPE, calldepth int, v ...any) *Logging {
	return static_lo.println(format, level, k1(calldepth), v...)
}

func fprintln(format *string, _format _FORMAT, level, stacktrace LEVELTYPE, calldepth int, formatter *string, attrFormat *AttrFormat, v ...any) {
	var bs []byte
	if format == nil {
		bs = fmt.Append([]byte{}, v...)
	} else {
		bs = fmt.Appendf([]byte{}, *format, v...)
	}
	consolewrite(bs, level, stacktrace, _format, k1(calldepth), formatter, attrFormat)
}

func getlevelname(level LEVELTYPE) (levelname []byte) {
	switch level {
	case LEVEL_ALL:
		levelname = []byte("ALL")
	case LEVEL_DEBUG:
		levelname = _DEBUG
	case LEVEL_INFO:
		levelname = _INFO
	case LEVEL_WARN:
		levelname = _WARN
	case LEVEL_ERROR:
		levelname = _ERROR
	case LEVEL_FATAL:
		levelname = _FATALE
	default:
		levelname = []byte{}
	}
	return
}

// Logging is the primary data structure for configuring and managing logging behavior.
type Logging struct {
	_level        LEVELTYPE                 // Log level, e.g., DEBUG, INFO, WARN, ERROR, etc.
	_format       _FORMAT                   // Log format.
	_rwLock       *sync.RWMutex             // Read-write lock for concurrent safe access to the logging struct.
	_fileDir      string                    // Directory path where log files are stored.
	_fileName     string                    // Base name of the log file.
	_maxSize      int64                     // Maximum size of a single log file.
	_unit         _UNIT                     // Size unit, e.g., Byte, KB, MB, etc.
	_cutmode      _CUTMODE                  // Log file cutting mode, e.g., by size or by time.
	_mode         _MODE_TIME                // Time-based rolling mode for log files, e.g., daily, weekly, etc.
	_filehandler  *fileHandler              // File handler for operations on log files.
	_isFileWell   bool                      // Indicates whether the log file is in good condition.
	_formatter    string                    // Formatting string for customizing the log output format.
	_maxBackup    int                       // Maximum number of backup log files to keep.
	_isConsole    bool                      // Whether to also output logs to the console.
	_gzip         bool                      // Whether to enable GZIP compression for old log files.
	prevTime      int64                     // The timestamp of last print
	callDepth     int                       // the depth of function call
	stacktrace    LEVELTYPE                 // Log level, e.g., DEBUG, INFO, WARN, ERROR, etc.
	customHandler func(lc *LogContext) bool // Custom log handler function allowing users to define additional log processing logic.
	atStart       atomic.Int32
	atStop        atomic.Int32
	leveloption   [5]*LevelOption
	attrFormat    *AttrFormat
	tmTimer       *time.Timer
	err           error
}

// NewLogger creates and returns a new instance of the Logging struct.
// This function initializes a Logging object with default values or specific configurations as needed.
func NewLogger() (log *Logging) {
	log = &Logging{_level: default_level, _cutmode: _TIMEMODE, _rwLock: new(sync.RWMutex), _format: default_format, _isConsole: true}
	log.newfileHandler()
	return
}

// SetConsole sets the flag to determine whether log messages should also be output to the console.
// This method modifies the _isConsole field of the Logging struct and returns a pointer to the Logging instance for method chaining.
func (t *Logging) SetConsole(_isConsole bool) *Logging {
	t._isConsole = _isConsole
	return t
}

// Debug logs a message at the DEBUG level.
// Accepts any number of arguments to format the log entry.
// Returns the Logging instance for method chaining.
//
// Parameters:
//   - v: Variadic arguments to be logged.
//
// Returns:
//   - *Logging: The current Logging instance for chaining.
func (t *Logging) Debug(v ...any) *Logging {
	return t.println(nil, LEVEL_DEBUG, 2, v...)
}

// Info logs a message at the INFO level.
// Accepts any number of arguments to format the log entry.
// Returns the Logging instance for method chaining.
//
// Parameters:
//   - v: Variadic arguments to be logged.
//
// Returns:
//   - *Logging: The current Logging instance for chaining.
func (t *Logging) Info(v ...any) *Logging {
	return t.println(nil, LEVEL_INFO, 2, v...)
}

// Warn logs a message at the WARN level.
// Accepts any number of arguments to format the log entry.
// Returns the Logging instance for method chaining.
//
// Parameters:
//   - v: Variadic arguments to be logged.
//
// Returns:
//   - *Logging: The current Logging instance for chaining.
func (t *Logging) Warn(v ...any) *Logging {
	return t.println(nil, LEVEL_WARN, 2, v...)
}

// Error logs a message at the ERROR level.
// Accepts any number of arguments to format the log entry.
// Returns the Logging instance for method chaining.
//
// Parameters:
//   - v: Variadic arguments to be logged.
//
// Returns:
//   - *Logging: The current Logging instance for chaining.
func (t *Logging) Error(v ...any) *Logging {
	return t.println(nil, LEVEL_ERROR, 2, v...)
}

// Fatal logs a message at the FATAL level and may terminate the application.
// Accepts any number of arguments to format the log entry.
// Returns the Logging instance for method chaining.
//
// Parameters:
//   - v: Variadic arguments to be logged.
//
// Returns:
//   - *Logging: The current Logging instance for chaining.
func (t *Logging) Fatal(v ...any) *Logging {
	return t.println(nil, LEVEL_FATAL, 2, v...)
}

// Debugf logs a formatted message at the DEBUG level.
// Takes a format string followed by variadic arguments to be formatted.
// Returns the Logging instance for method chaining.
//
// Parameters:
//   - format: The format string for the log entry.
//   - v: Variadic arguments to be formatted according to the format string.
//
// Returns:
//   - *Logging: The current Logging instance for chaining.
func (t *Logging) Debugf(format string, v ...any) *Logging {
	return t.println(&format, LEVEL_DEBUG, 2, v...)
}

// Infof logs a formatted message at the INFO level.
// Takes a format string followed by variadic arguments to be formatted.
// Returns the Logging instance for method chaining.
//
// Parameters:
//   - format: The format string for the log entry.
//   - v: Variadic arguments to be formatted according to the format string.
//
// Returns:
//   - *Logging: The current Logging instance for chaining.
func (t *Logging) Infof(format string, v ...any) *Logging {
	return t.println(&format, LEVEL_INFO, 2, v...)
}

// Warnf logs a formatted message at the WARN level.
// Takes a format string followed by variadic arguments to be formatted.
// Returns the Logging instance for method chaining.
//
// Parameters:
//   - format: The format string for the log entry.
//   - v: Variadic arguments to be formatted according to the format string.
//
// Returns:
//   - *Logging: The current Logging instance for chaining.
func (t *Logging) Warnf(format string, v ...any) *Logging {
	return t.println(&format, LEVEL_WARN, 2, v...)
}

// Errorf logs a formatted message at the ERROR level.
// Takes a format string followed by variadic arguments to be formatted.
// Returns the Logging instance for method chaining.
//
// Parameters:
//   - format: The format string for the log entry.
//   - v: Variadic arguments to be formatted according to the format string.
//
// Returns:
//   - *Logging: The current Logging instance for chaining.
func (t *Logging) Errorf(format string, v ...any) *Logging {
	return t.println(&format, LEVEL_ERROR, 2, v...)
}

// Fatalf logs a formatted message at the FATAL level and may terminate the application.
// Takes a format string followed by variadic arguments to be formatted.
// Returns the Logging instance for method chaining.
//
// Parameters:
//   - format: The format string for the log entry.
//   - v: Variadic arguments to be formatted according to the format string.
//
// Returns:
//   - *Logging: The current Logging instance for chaining.
func (t *Logging) Fatalf(format string, v ...any) *Logging {
	return t.println(&format, LEVEL_FATAL, 2, v...)
}

func (t *Logging) WriteBin(bs []byte) (bakfn string, err error) {
	if t._isFileWell {
		var openFileErr error
		if t._filehandler.mustBackUp(len(bs)) {
			bakfn, err, openFileErr = t.backUp()
		}
		if openFileErr == nil {
			t._rwLock.RLock()
			_, err = t._filehandler.write(bs)
			t._rwLock.RUnlock()
		}
	} else {
		err = errors.New("no log file found")
	}
	return
}
func (t *Logging) Write(bs []byte) (n int, err error) {
	if t._isFileWell {
		var openFileErr error
		if t._filehandler.mustBackUp(len(bs)) {
			_, err, openFileErr = t.backUp()
		}
		if openFileErr == nil {
			t._rwLock.RLock()
			n, err = t._filehandler.write(bs)
			t._rwLock.RUnlock()
		}
	} else {
		err = errors.New("no log file found")
	}
	return
}

// SetFormat sets the logging format to the specified format type.
//
// Parameters:
//   - format: The desired format for log entries, represented by the _FORMAT type.
//
// Returns:
//   - *Logging: A Logging instance to enable method chaining.
func (t *Logging) SetFormat(format _FORMAT) *Logging {
	t._format = format
	return t
}

// SetLevel sets the logging level to the specified level type.
//
// Parameters:
//   - level: The logging level (e.g., DEBUG, INFO, WARN), represented by the LEVELTYPE type.
//
// Returns:
//   - *Logging: A Logging instance to enable method chaining.
func (t *Logging) SetLevel(level LEVELTYPE) *Logging {
	t._level = level
	return t
}

// SetFormatter specifies a custom format string for the logging output.
//
//	default:  "{level}{time} {file} {message}\n"
//
// Parameters:
//   - formatter: A string defining the format for log entries, allowing custom log entry layouts.
//
// Returns:
//   - *Logging: A Logging instance to enable method chaining.
func (t *Logging) SetFormatter(formatter string) *Logging {
	t._formatter = formatter
	return t
}

// SetRollingFile
//
// Use SetOption() instead.
// 按日志文件大小分割日志文件
// fileDir 日志文件夹路径
// fileName 日志文件名
// maxFileSize  日志文件大小最大值
// unit    日志文件大小单位
func (t *Logging) SetRollingFile(fileDir, fileName string, maxFileSize int64, unit _UNIT) (l *Logging, err error) {
	return t.SetRollingFileLoop(fileDir, fileName, maxFileSize, unit, 0)
}

// SetRollingFileLoop
//
// Use SetOption() instead.
// 按日志文件大小分割日志文件，指定保留的最大日志文件数
// fileDir 日志文件夹路径
// fileName 日志文件名
// maxFileSize  日志文件大小最大值
// unit    	日志文件大小单位
// maxFileNum  留的日志文件数
func (t *Logging) SetRollingFileLoop(fileDir, fileName string, maxFileSize int64, unit _UNIT, maxBackup int) (l *Logging, err error) {
	t._rwLock.Lock()
	defer t._rwLock.Unlock()
	if fileDir == "" {
		fileDir, _ = os.Getwd()
	}
	t._fileDir, t._fileName, t._maxSize, t._maxBackup, t._unit = fileDir, fileName, maxFileSize, maxBackup, unit
	t._cutmode = _SIZEMODE
	if t._filehandler != nil {
		t._filehandler.close()
	}
	t.newfileHandler()
	if err = t._filehandler.openFileHandler(); err == nil {
		t._isFileWell = true
	}
	t.err = err
	return t, err
}

// SetRollingDaily
//
// Use SetOption() instead.
// 按日期分割日志文件
// fileDir 日志文件夹路径
// fileName 日志文件名
func (t *Logging) SetRollingDaily(fileDir, fileName string) (*Logging, error) {
	return t.SetRollingByTime(fileDir, fileName, MODE_DAY)
}

// SetRollingByTime
//
// 指定按 小时，天，月 分割日志文件
// fileDir 日志文件夹路径
// fileName 日志文件名
// mode   指定 小时，天，月
func (t *Logging) SetRollingByTime(fileDir, fileName string, mode _MODE_TIME) (l *Logging, err error) {
	t._rwLock.Lock()
	defer t._rwLock.Unlock()
	if fileDir == "" {
		fileDir, _ = os.Getwd()
	}
	t._fileDir, t._fileName, t._mode = fileDir, fileName, mode
	t._cutmode = _TIMEMODE
	if t._filehandler != nil {
		t._filehandler.close()
	}
	t.newfileHandler()
	if err = t._filehandler.openFileHandler(); err == nil {
		t._isFileWell = true
		t.zeroTimer()
	}
	t.err = err
	return t, err
}

// SetGzipOn
//
// Use SetOption() instead.
func (t *Logging) SetGzipOn(is bool) *Logging {
	t._gzip = is
	if t._filehandler != nil {
		t._filehandler._gzip = is
	}
	return t
}

// SetOption applies the configuration options specified in the Option struct to the Logging instance.
// This method updates the fields of the Logging struct according to the provided Option and returns a pointer to the Logging instance for method chaining.
//
// e.g.
//
// SetOption(&Option{Level: LEVEL_DEBUG, Console: true, FileOption: &FileSizeMode{Filename: "test.log", Maxsize: 500, Maxbackup: 3, IsCompress: false}})
func (t *Logging) SetOption(option *Option) *Logging {
	t._rwLock.Lock()
	defer t._rwLock.Unlock()
	t.getOptionArgs(option)
	if option.FileOption != nil {
		if abspath, err := filepath.Abs(option.FileOption.FilePath()); err == nil {
			t._fileDir = filepath.Dir(abspath)
		} else {
			fprintln(nil, default_format, LEVEL_ERROR, 0, 1, nil, nil, err.Error())
			t._fileDir, _ = os.Getwd()
		}
		t._fileName = filepath.Base(option.FileOption.FilePath())
		if t._filehandler != nil {
			t._filehandler.close()
		}
		t.newfileHandler()
		if err := t._filehandler.openFileHandler(); err == nil {
			t._isFileWell = true
		} else {
			fprintln(nil, default_format, LEVEL_ERROR, 0, 1, nil, nil, err.Error())
			t.err = err
		}

		if t._cutmode&_TIMEMODE == _TIMEMODE {
			t.zeroTimer()
		}
	}
	return t
}

func (t *Logging) getOptionArgs(option *Option) {
	if option.Formatter != "" {
		t._formatter = option.Formatter
	}
	if option.AttrFormat != nil {
		t.attrFormat = option.AttrFormat
	}
	if option.Format != 0 {
		t._format = option.Format
	}
	t._isConsole = option.Console
	t.callDepth = option.CallDepth
	t.customHandler = option.CustomHandler
	t.stacktrace = option.Stacktrace
	t._level = option.Level
	if option.FileOption != nil {
		t._cutmode = option.FileOption.Cutmode()
		if t._cutmode != _TIMEMODE && t._cutmode != _SIZEMODE && t._cutmode != _MIXEDMODE {
			t._cutmode = _MIXEDMODE
		}
		t._maxBackup, t._gzip = option.FileOption.MaxBackup(), option.FileOption.Compress()
		if t._cutmode&_SIZEMODE == _SIZEMODE {
			t._maxSize, t._unit = option.FileOption.MaxSize(), 1
			if t._maxSize <= 0 {
				t._maxSize = 1 << 30
			}
		}
		if t._cutmode&_TIMEMODE == _TIMEMODE {
			t._mode = option.FileOption.TimeMode()
			if t._mode != MODE_DAY && t._mode != MODE_HOUR && t._mode != MODE_MONTH {
				t._mode = MODE_DAY
			}
		}
	}
}

func (t *Logging) newfileHandler() {
	t._filehandler = new(fileHandler)
	t._filehandler.logger, t._filehandler._fileDir, t._filehandler._fileName, t._filehandler._maxSize, t._filehandler._cutmode, t._filehandler._unit, t._filehandler._maxbackup, t._filehandler._mode, t._filehandler._gzip = t, t._fileDir, t._fileName, t._maxSize, t._cutmode, t._unit, t._maxBackup, t._mode, t._gzip
}

func (t *Logging) backUp() (bakfn string, err, openFileErr error) {
	t._rwLock.Lock()
	defer t._rwLock.Unlock()
	if !t._filehandler.mustBackUp(0) {
		return
	}
	if err = t._filehandler.close(); err != nil {
		fprintln(nil, t._format, LEVEL_ERROR, t.stacktrace, 1, nil, nil, err.Error())
		return
	}

	for i := 0; i < 16; i++ {
		if bakfn, err = t._filehandler.rename(); err != nil {
			fprintln(nil, t._format, LEVEL_ERROR, t.stacktrace, 1, nil, nil, err.Error())
			<-time.After(time.Millisecond)
		} else {
			break
		}
	}

	if openFileErr = t._filehandler.openFileHandler(); openFileErr != nil {
		fprintln(nil, t._format, LEVEL_ERROR, t.stacktrace, 1, nil, nil, openFileErr.Error())
		t.err = openFileErr
	}
	return
}

func (t *Logging) println(format *string, _level LEVELTYPE, calldepth int, v ...any) *Logging {
	if t._level > _level {
		return t
	}
	if t.err != nil {
		return t
	}
	if t.customHandler != nil && !t.customHandler(&LogContext{Level: _level, Args: v}) {
		return t
	}
	var buf *buffer.Buffer
	defer func() {
		if buf != nil {
			buf.Free()
		}
	}()
	var bs []byte
	if t.callDepth > 0 {
		calldepth += t.callDepth
	}

	if t._isFileWell {
		if t._format != FORMAT_NANO {
			if format == nil {
				bs = fmt.Append([]byte{}, v...)
			} else {
				bs = fmt.Appendf([]byte{}, *format, v...)
			}
			if ol := t.leveloption[_level-1]; ol != nil {
				buf = getOutBuffer(bs, _level, ol.Format, k1(calldepth), &ol.Formatter, t.stacktrace, t.attrFormat)
			} else {
				buf = getOutBuffer(bs, _level, t._format, k1(calldepth), &t._formatter, t.stacktrace, t.attrFormat)
			}
			if t.attrFormat != nil && t.attrFormat.SetBodyFmt != nil {
				bs = t.attrFormat.SetBodyFmt(_level, buf.Bytes())
			} else {
				bs = buf.Bytes()
			}
		} else {
			if format == nil {
				bs = fmt.Appendln([]byte{}, v...)
			} else {
				bs = fmt.Appendf([]byte{}, *format+"\n", v...)
			}
		}
		var openFileErr error
		if t._filehandler.mustBackUp(len(bs)) {
			_, openFileErr, _ = t.backUp()
		}
		if openFileErr == nil {
			t._rwLock.RLock()
			t._filehandler.write(bs)
			t._rwLock.RUnlock()
		}
	}
	if t._isConsole {
		if bs != nil {
			consolewriter(bs, false)
		} else {
			if ol := t.leveloption[_level-1]; ol != nil {
				fprintln(format, ol.Format, _level, t.stacktrace, k1(calldepth), &ol.Formatter, t.attrFormat, v...)
			} else {
				fprintln(format, t._format, _level, t.stacktrace, k1(calldepth), &t._formatter, t.attrFormat, v...)
			}
		}
	}
	return t
}

func SetLevelOption(level LEVELTYPE, option *LevelOption) *Logging {
	return static_lo.SetLevelOption(level, option)
}

func (t *Logging) SetLevelOption(level LEVELTYPE, option *LevelOption) *Logging {
	if level > LEVEL_ALL && level < LEVEL_OFF {
		t.leveloption[level-1] = option
	}
	return t
}

type fileHandler struct {
	logger     *Logging
	fileHandle File
	_fileDir   string
	_fileName  string
	file       *os.File
	_maxSize   int64
	_fileSize  int64
	_fileSize2 int64
	_lastPrint int64
	_prevPrint int64
	_unit      _UNIT
	_cutmode   _CUTMODE
	_maxbackup int
	_gzip      bool
	_mode      _MODE_TIME
}

func (t *fileHandler) openFileHandler() (e error) {
	if t._fileDir == "" || t._fileName == "" {
		e = errors.New("log filePath is null or error")
		return
	}
	e = mkdirAll(t._fileDir)
	if e != nil {
		return
	}
	fname := filepath.Join(t._fileDir, t._fileName)
	if t.file, e = os.OpenFile(fname, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666); e == nil {
		t.fileHandle, e = New(t.file)
	}
	if e != nil {
		fprintln(nil, default_format, LEVEL_ERROR, 0, 1, nil, nil, e.Error())
		return
	}
	if fs, err := t.file.Stat(); err == nil {
		t._fileSize = fs.Size()
		atomic.StoreInt64(&t._fileSize2, t._fileSize)
		t._lastPrint = fs.ModTime().Unix()
	} else {
		e = err
	}
	return
}

func (t *fileHandler) addFileSize(size int64) {
	atomic.AddInt64(&t._fileSize, size)
}

func (t *fileHandler) write(bs []byte) (n int, e error) {
	defer recoverable(&e)
	if bs != nil {
		if n, e = t.fileHandle.Write(bs); e == nil {
			if n > 0 {
				t.addFileSize(int64(n))
			}
			if t._cutmode&_TIMEMODE == _TIMEMODE {
				t._lastPrint = loctime().Unix()
				if t._prevPrint > 0 && t._lastPrint-t._prevPrint > 2 && t.logger.tmTimer == nil {
					t.logger.zeroTimer()
				} else if t.logger.tmTimer != nil {
					t.logger.zeroTimerStop()
				}
				t._prevPrint = t._lastPrint
			}
		}
	}
	return
}

func (t *fileHandler) mustBackUp(addsize int) bool {
	if t._fileSize == 0 {
		return false
	}
	if t._cutmode&_TIMEMODE == _TIMEMODE {
		if t._lastPrint > 0 && !isCurrentTime(t._mode, t._lastPrint) {
			return true
		}
	}
	if t._cutmode&_SIZEMODE == _SIZEMODE {
		if addsize > 0 {
			if atomic.AddInt64(&t._fileSize2, int64(addsize)) >= t._maxSize*int64(t._unit) {
				return true
			}
		}
		return t._fileSize > 0 && atomic.LoadInt64(&t._fileSize) >= t._maxSize*int64(t._unit)
	}
	return false
}

func (t *fileHandler) rename() (bckupfilename string, err error) {
	if t._cutmode&_TIMEMODE == _TIMEMODE {
		bckupfilename = getBackupDayliFileName(t._lastPrint, t._fileDir, t._fileName, t._mode, t._gzip)
	} else {
		bckupfilename, err = getBackupRollFileName(t._fileDir, t._fileName, t._gzip)
	}
	if bckupfilename != "" && err == nil {
		oldPath := filepath.Join(t._fileDir, t._fileName)
		newPath := filepath.Join(t._fileDir, bckupfilename)
		if err = os.Rename(oldPath, newPath); err == nil {
			go func() {
				defer recoverable(nil)
				if t._gzip {
					if err = lgzip(newPath+".gz", bckupfilename, newPath); err == nil {
						os.Remove(newPath)
					}
				}
				if t._maxbackup > 0 {
					maxbackup(t._fileDir, t._fileName, t._maxbackup)
				}
			}()
		}
	}
	return
}

func (t *fileHandler) close() (err error) {
	defer recoverable(&err)
	if t.fileHandle != nil {
		err = t.fileHandle.Close()
	} else if t.file != nil {
		err = t.file.Close()
	}
	return
}

//func tomorSecond(mode _MODE_TIME) int64 {
//	now := loctime()
//	switch mode {
//	case MODE_DAY:
//		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()).Unix()
//	case MODE_HOUR:
//		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location()).Unix()
//	case MODE_MONTH:
//		return time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1).Unix()
//	default:
//		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()).Unix()
//	}
//}

func isCurrentTime(mode _MODE_TIME, timestamp int64) bool {
	now := loctime()
	switch mode {
	case MODE_DAY:
		return timestamp >= time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	case MODE_HOUR:
		return timestamp >= time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location()).Unix()
	case MODE_MONTH:
		return timestamp >= time.Date(now.Year(), now.Month(), 0, 0, 0, 0, 0, now.Location()).Unix()
	}
	return false
}

func backupStr4Time(mode _MODE_TIME, now time.Time) string {
	switch mode {
	case MODE_HOUR:
		return now.Format(_DATEFORMAT_HOUR)
	case MODE_MONTH:
		return now.Format(_DATEFORMAT_MONTH)
	default:
		return now.Format(_DATEFORMAT_DAY)
	}
}

//func _yestStr(mode _MODE_TIME, now time.Time) string {
//	//now := loctime()
//	switch mode {
//	case MODE_DAY:
//		return now.AddDate(0, 0, -1).Format(_DATEFORMAT_DAY)
//	case MODE_HOUR:
//		return now.Add(-1 * time.Hour).Format(_DATEFORMAT_HOUR)
//	case MODE_MONTH:
//		return now.AddDate(0, -1, 0).Format(_DATEFORMAT_MONTH)
//	default:
//		return now.AddDate(0, 0, -1).Format(_DATEFORMAT_DAY)
//	}
//}

func getBackupDayliFileName(unixTimestamp int64, dir, filename string, mode _MODE_TIME, isGzip bool) (bckupfilename string) {
	timeStr := backupStr4Time(mode, time.Unix(unixTimestamp, 0))
	index := strings.LastIndex(filename, ".")
	if index <= 0 {
		index = len(filename)
	}
	fname := filename[:index]
	suffix := filename[index:]
	bckupfilename = fmt.Sprint(fname, "_", timeStr, suffix)
	if isGzip {
		if isFileExist(fmt.Sprint(filepath.Join(dir, bckupfilename), ".gz")) {
			bckupfilename = _getBackupfilename(1, dir, fmt.Sprint(fname, "_", timeStr), suffix, isGzip)
		}
	} else {
		if isFileExist(fmt.Sprint(filepath.Join(dir, bckupfilename))) {
			bckupfilename = _getBackupfilename(1, dir, fmt.Sprint(fname, "_", timeStr), suffix, isGzip)
		}
	}
	return
}

func _getDirList(dir string) ([]os.DirEntry, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.ReadDir(-1)
}

func getBackupRollFileName(dir, filename string, isGzip bool) (bckupfilename string, er error) {
	list, err := _getDirList(dir)
	if err != nil {
		er = err
		return
	}
	index := strings.LastIndex(filename, ".")
	if index <= 0 {
		index = len(filename)
	}
	fname := filename[:index]
	suffix := filename[index:]
	i := 1
	for _, fd := range list {
		pattern := fmt.Sprint(`^`, fname, `_[\d]{1,}`, suffix, `$`)
		if isGzip {
			pattern = fmt.Sprint(`^`, fname, `_[\d]{1,}`, suffix, `.gz$`)
		}
		if matchString(pattern, fd.Name()) {
			i++
		}
	}
	bckupfilename = _getBackupfilename(i, dir, fname, suffix, isGzip)
	return
}

func _getBackupfilename(count int, dir, filename, suffix string, isGzip bool) (bckupfilename string) {
	bckupfilename = fmt.Sprint(filename, "_", count, suffix)
	if isGzip {
		if isFileExist(fmt.Sprint(filepath.Join(dir, bckupfilename), ".gz")) {
			return _getBackupfilename(count+1, dir, filename, suffix, isGzip)
		}
	} else {
		if isFileExist(fmt.Sprint(filepath.Join(dir, bckupfilename))) {
			return _getBackupfilename(count+1, dir, filename, suffix, isGzip)
		}
	}
	return
}

func consolewrite(s []byte, level, stacktrace LEVELTYPE, flag _FORMAT, calldepth int, formatter *string, attrFormat *AttrFormat) {
	if flag != FORMAT_NANO {
		buf := getOutBuffer(s, level, flag, k1(calldepth), formatter, stacktrace, attrFormat)
		defer buf.Free()
		if attrFormat != nil && attrFormat.SetBodyFmt != nil {
			consolewriter(attrFormat.SetBodyFmt(level, buf.Bytes()), false)
		} else {
			consolewriter(buf.Bytes(), false)
		}
	} else {
		if attrFormat != nil && attrFormat.SetBodyFmt != nil {
			consolewriter(attrFormat.SetBodyFmt(level, s), false)
		} else {
			consolewriter(s, true)
		}
	}
}

func consolewriter(bs []byte, newline bool) {
	if newline {
		os.Stdout.Write(append(bs, '\n'))
	} else {
		os.Stdout.Write(bs)
	}
}

func k1(calldepth int) int {
	return calldepth + 1
}

func getOutBuffer(s []byte, level LEVELTYPE, format _FORMAT, calldepth int, formatter *string, stacktrace LEVELTYPE, attrFormat *AttrFormat) *buffer.Buffer {
	return output(format, k1(calldepth), s, level, formatter, stacktrace, attrFormat)
}

func mkdirAll(dir string) (e error) {
	_, er := os.Stat(dir)
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			if os.IsPermission(err) {
				e = err
			}
		}
	}
	return
}

func maxbackup(dir, filename string, maxcount int) {
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]
	if entries, err := os.ReadDir(dir); err == nil {
		if len(entries) > maxcount {
			sort.Slice(entries, func(i, j int) bool {
				f1, _ := entries[i].Info()
				f2, _ := entries[j].Info()
				return f1.ModTime().Unix() < f2.ModTime().Unix()
			})
			rms := make([]string, 0)
			for _, entry := range entries {
				if !entry.IsDir() {
					parrent := fmt.Sprint("^", name, "(_\\d+){0,}", "_\\d+", ext, "(\\.gz){0,}$")
					if matchString(parrent, entry.Name()) {
						filePath := filepath.Join(dir, entry.Name())
						rms = append(rms, filePath)
					}
				}
			}
			if len(rms) > maxcount {
				for i := 0; i < len(rms)-maxcount; i++ {
					os.Remove(rms[i])
				}
			}
		}
	}
}

func isFileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func recoverable(err *error) {
	if e := recover(); e != nil {
		if err != nil {
			*err = fmt.Errorf("%v", e)
		}
	}
}

func matchString(pattern string, s string) bool {
	b, err := regexp.MatchString(pattern, s)
	if err != nil {
		b = false
	}
	return b
}

func loctime() time.Time {
	if TIME_DEVIATION != 0 {
		return time.Now().Add(TIME_DEVIATION)
	} else {
		return time.Now()
	}
}

func lgzip(gzfile, gzname, srcfile string) (err error) {
	var gf *os.File
	if gf, err = os.Create(gzfile); err == nil {
		defer gf.Close()
		var f1 *os.File
		if f1, err = os.Open(srcfile); err == nil {
			defer f1.Close()
			gw := gzip.NewWriter(gf)
			defer gw.Close()
			gw.Header.Name = gzname
			var buf bytes.Buffer
			io.Copy(&buf, f1)
			_, err = gw.Write(buf.Bytes())
		}
	}
	return
}

var m = hashmap.NewLimitHashMap[uintptr, runtime.Frame](1 << 13)

func output(flag _FORMAT, calldepth int, s []byte, level LEVELTYPE, formatter *string, stacktrace LEVELTYPE, attrFormat *AttrFormat) (buf *buffer.Buffer) {
	var callstack *callStack
	if flag&(FORMAT_SHORTFILENAME|FORMAT_LONGFILENAME|FORMAT_RELATIVEFILENAME) != 0 {
		callstack = collectCallStack(k1(calldepth), flag&FORMAT_FUNC != 0, callstack, stacktrace > LEVEL_ALL && stacktrace <= level)
	}
	return formatmsg(s, loctime(), callstack, flag, level, formatter, attrFormat)
}

func formatmsg(msg []byte, t time.Time, callstack *callStack, flag _FORMAT, level LEVELTYPE, formatter *string, attrFormat *AttrFormat) (buf *buffer.Buffer) {
	buf = buffer.NewBufferByPool()
	var levelbuf, timebuf, filebuf *buffer.Buffer
	is_default_formatter := formatter == nil || *formatter == ""
	if is_default_formatter {
		levelbuf, timebuf, filebuf = buf, buf, buf
	} else {
		levelbuf = buffer.NewBufferWithCapacity(7)
		timebuf = buffer.NewBufferWithCapacity(20)
		filebuf = buffer.NewBuffer()
	}
	if flag&FORMAT_LEVELFLAG != 0 {
		if attrFormat != nil && attrFormat.SetLevelFmt != nil {
			levelbuf.WriteString(attrFormat.SetLevelFmt(level))
		} else {
			levelbuf.Write(getlevelname(level))
		}
	}
	if flag&timeFlags != 0 {
		if attrFormat != nil && attrFormat.SetTimeFmt != nil {
			datestr, timestr, microsecond := attrFormat.SetTimeFmt()
			if flag&FORMAT_DATE != 0 && datestr != "" {
				timebuf.WriteString(datestr)
			}
			if flag&(FORMAT_TIME|FORMAT_MICROSECONDS) != 0 {
				timebuf.WriteString(timestr)
				if flag&FORMAT_MICROSECONDS != 0 {
					timebuf.WriteString(microsecond)
				}
			}
		} else {
			if flag&FORMAT_DATE != 0 {
				year, month, day := t.Date()
				timebuf.Write(itoa(year, 4))
				timebuf.WriteByte('/')
				timebuf.Write(itoa(int(month), 2))
				timebuf.WriteByte('/')
				timebuf.Write(itoa(day, 2))
				timebuf.WriteByte(' ')
			}
			if flag&(FORMAT_TIME|FORMAT_MICROSECONDS) != 0 {
				hour, min, sec := t.Clock()
				timebuf.Write(itoa(hour, 2))
				timebuf.WriteByte(':')
				timebuf.Write(itoa(min, 2))
				timebuf.WriteByte(':')
				timebuf.Write(itoa(sec, 2))
				if flag&FORMAT_MICROSECONDS != 0 {
					timebuf.WriteByte('.')
					timebuf.Write(itoa(t.Nanosecond()/1e3, 6))
				}
			}
		}
		if is_default_formatter {
			timebuf.WriteByte(' ')
		}
	}
	if flag&fileFlags != 0 {
		if callstack != nil {
			callstack.Pop(flag, filebuf)
			callStackPool.Put(&callstack)
		}
		if is_default_formatter {
			filebuf.WriteByte(' ')
		}
	}
	if is_default_formatter {
		buf.Write(msg)
		buf.WriteByte('\n')
	} else {
		parseAndFormatLog(formatter, buf, levelbuf, timebuf, filebuf, msg)
	}
	return
}

func parseAndFormatLog(formatStr *string, buf, levelbuf, timebuf, filebuf *buffer.Buffer, msg []byte) {
	if formatStr == nil || *formatStr == "" {
		buf.Write(msg)
		return
	}
	inPlaceholder := false
	placeholder := ""
	for _, c := range *formatStr {
		if inPlaceholder {
			if c == '}' {
				inPlaceholder = false
				switch placeholder {
				case "level":
					buf.Write(levelbuf.Bytes())
				case "time":
					buf.Write(timebuf.Bytes())
				case "file":
					buf.Write(filebuf.Bytes())
				case "message":
					buf.Write(msg)
				}
				placeholder = ""
			} else {
				placeholder += string(c)
			}
		} else if c == '{' {
			inPlaceholder = true
		} else {
			buf.WriteByte(byte(c))
		}
	}
}

func itoa(i int, wid int) []byte {
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	b[bp] = byte('0' + i)
	return b[bp:]
}

func (t *Logging) zeroTimerStop() {
	if t.atStop.CompareAndSwap(0, 1) {
		defer t.atStop.Store(0)
		if t.tmTimer != nil {
			t.tmTimer.Stop()
			t.tmTimer = nil
		}
	}
}

func (t *Logging) zeroTimer() {
	if t.atStart.CompareAndSwap(0, 1) {
		defer t.atStart.Store(0)
		if t.tmTimer == nil {
			t.tmTimer = time.AfterFunc(timeUntilNextWholeHour(), t.zeroCheck)
		}
	}
}

func (t *Logging) zeroCheck() {
	defer recoverable(nil)
	if t._filehandler.mustBackUp(0) {
		t.backUp()
	}
	t.zeroTimer()
}

func timeUntilNextWholeHour() (r time.Duration) {
	now := time.Now()
	nextWholeHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 1, 0, now.Location())
	if r = nextWholeHour.Sub(now); r < time.Second {
		r = time.Second
	}
	return
}
