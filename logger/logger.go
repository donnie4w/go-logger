// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/donnie4w/gofer/buffer"
	"github.com/donnie4w/gofer/hashmap"
)

const (
	VERSION string = "0.25.1"
)

type _LEVEL int8
type _UNIT int64
type _MODE_TIME uint8
type _CUTMODE int //dailyRolling ,rollingFile
type _FORMAT int

const _DATEFORMAT_DAY = "20060102"
const _DATEFORMAT_HOUR = "2006010215"
const _DATEFORMAT_MONTH = "200601"
const default_format = FORMAT_LEVELFLAG | FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME
const default_level = LEVEL_ALL
const default_formatter = "{level}{time} {file}:{message}\n"

var static_mu = new(sync.Mutex)

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
	/*无其他格式，只打印日志内容*/ /*no format, Only log content is printed*/
	FORMAT_NANO       _FORMAT = 64

	/*长文件名(文件绝对路径)及行数*/ /*full file name and line number*/
	FORMAT_LONGFILENAME = _FORMAT(8)

	/*短文件名及行数*/          /*final file name element and line number*/
	FORMAT_SHORTFILENAME = _FORMAT(16)

	/*日期时间精确到天*/ /*the date in the local time zone: 2009/01/23*/
	FORMAT_DATE  = _FORMAT(1)

	/*时间精确到秒*/  /*the time in the local time zone: 01:23:23*/
	FORMAT_TIME = _FORMAT(2)

	/*时间精确到微秒*/        /*microsecond resolution: 01:23:23.123123.*/
	FORMAT_MICROSECNDS = _FORMAT(4)

	/*日志级别表示*/       /*Log level flag. e.g. [DEBUG],[INFO],[WARN],[ERROR],[FATAL]*/
	FORMAT_LEVELFLAG = _FORMAT(32)
)

const (
	/*日志级别：ALL 最低级别*/ /*Log level: LEVEL_ALL is the lowest level,If the log level is this level, logs of other levels can be printed*/
	LEVEL_ALL         _LEVEL = iota

	/*日志级别：DEBUG 小于INFO*/ /*Log level: ALL<DEBUG<INFO*/
	LEVEL_DEBUG

	/*日志级别：INFO 小于 WARN*/ /*Log level: DEBUG<INFO<WARN*/
	LEVEL_INFO

	/*日志级别：WARN 小于 ERROR*/ /*Log level: INFO<WARN<ERROR*/
	LEVEL_WARN

	/*日志级别：ERROR 小于 FATAL*/ /*Log level: WARN<ERROR<FATAL*/
	LEVEL_ERROR

	/*日志级别：FATAL 小于 OFF*/ /*Log level: ERROR<FATAL<OFF*/
	LEVEL_FATAL

	/*日志级别：off 不打印任何日志*/ /*Log level: LEVEL_OFF means none of the logs can be printed*/
	LEVEL_OFF
)

var DEBUGNAME, INFONAME, WARNNAME, ERRORNAME, FATALNAME = []byte("[DEBUG]"), []byte("[INFO]"), []byte("[WARN]"), []byte("[ERROR]"), []byte("[FATAL]")

const (
	_TIMEMODE _CUTMODE = 1
	_SIZEMODE _CUTMODE = 2
)

/*设置打印格式*/
func SetFormat(format _FORMAT) *Logging {
	return static_lo.SetFormat(format)
}

/*设置控制台日志级别，默认ALL*/
// Setting the log Level
func SetLevel(level _LEVEL) *Logging {
	return static_lo.SetLevel(level)
}

/*设置输出格式，默认: "{level}{time} {file}:{message}\n" */
func SetFormatter(formatter string) *Logging {
	return static_lo.SetFormatter(formatter)
}

// print logs on the console or not. default true
func SetConsole(on bool) *Logging {
	return static_lo.SetConsole(on)

}

/*获得全局Logger对象*/ /*return the default log object*/
func GetStaticLogger() *Logging {
	return _staticLogger()
}

// when the log file(fileDir+`\`+fileName) exceeds the specified size(maxFileSize), it will be backed up with a specified file name
// Parameters:
//   - fileDir   :directory where log files are stored, If it is the current directory, you also can set it to ""
//   - fileName  : log file name
//   - maxFileSize :  maximum size of a log file
//   - unit		   :  size unit :  KB,MB,GB,TB
//
// Deprecated
// Use SeOption() instead.
func SetRollingFile(fileDir, fileName string, maxFileSize int64, unit _UNIT) (l *Logging, err error) {
	return SetRollingFileLoop(fileDir, fileName, maxFileSize, unit, 0)
}

// yesterday's log data is backed up to a specified log file each day
// Parameters:
//   - fileDir   :directory where log files are stored, If it is the current directory, you also can set it to ""
//   - fileName  : log file name
//
// Deprecated
// Use SeOption() instead.
func SetRollingDaily(fileDir, fileName string) (l *Logging, err error) {
	return SetRollingByTime(fileDir, fileName, MODE_DAY)
}

// like SetRollingFile,but only keep (maxFileNum) current files
// - maxFileNum : the number of files that are retained
// Deprecated
// Use SeOption() instead.
func SetRollingFileLoop(fileDir, fileName string, maxFileSize int64, unit _UNIT, maxFileNum int) (l *Logging, err error) {
	return static_lo.SetRollingFileLoop(fileDir, fileName, maxFileSize, unit, maxFileNum)
}

// like SetRollingDaily,but supporte hourly backup ,dayly backup and monthly backup
// mode : 	MODE_HOUR    MODE_DAY   MODE_MONTH
// Deprecated
// Use SeOption() instead.
func SetRollingByTime(fileDir, fileName string, mode _MODE_TIME) (l *Logging, err error) {
	return static_lo.SetRollingByTime(fileDir, fileName, mode)
}

// when set true, the specified backup file of both SetRollingFile and SetRollingFileLoop will be save as a compressed file
// Deprecated
// Use SeOption() instead.
func SetGzipOn(is bool) (l *Logging) {
	return static_lo.SetGzipOn(is)
}

// 配置对象
func SetOption(option *Option) *Logging {
	return static_lo.SetOption(option)
}

func _staticLogger() *Logging {
	return static_lo
}

// Logs are printed at the DEBUG level
func Debug(v ...interface{}) *Logging {
	_println(LEVEL_DEBUG, default_level, 2, v...)
	return _staticLogger()
}

// Logs are printed at the INFO level
func Info(v ...interface{}) *Logging {
	_println(LEVEL_INFO, default_level, 2, v...)
	return _staticLogger()
}

// Logs are printed at the WARN level
func Warn(v ...interface{}) *Logging {
	_println(LEVEL_WARN, default_level, 2, v...)
	return _staticLogger()
}

// Logs are printed at the ERROR level
func Error(v ...interface{}) *Logging {
	_println(LEVEL_ERROR, default_level, 2, v...)
	return _staticLogger()
}

// Logs are printed at the FATAL level
func Fatal(v ...interface{}) *Logging {
	_println(LEVEL_FATAL, default_level, 2, v...)
	return _staticLogger()
}

func _println(level, _default_level _LEVEL, calldepth int, v ...interface{}) {
	_staticLogger().println(level, k1(calldepth), v...)
}

func fprintln(_format _FORMAT, level, _ _LEVEL, calldepth int, formatter *string, v ...interface{}) {
	_console(fmt.Append([]byte{}, v...), getlevelname(level, default_format), _format, k1(calldepth), formatter)
}

func getlevelname(level _LEVEL, format _FORMAT) (levelname []byte) {
	if format == FORMAT_NANO {
		return
	}
	switch level {
	case LEVEL_ALL:
		levelname = []byte("ALL")
	case LEVEL_DEBUG:
		levelname = DEBUGNAME
	case LEVEL_INFO:
		levelname = INFONAME
	case LEVEL_WARN:
		levelname = WARNNAME
	case LEVEL_ERROR:
		levelname = ERRORNAME
	case LEVEL_FATAL:
		levelname = FATALNAME
	default:
		levelname = []byte("")
	}
	return
}

/*————————————————————————————————————————————————————————————————————————————*/
type Logging struct {
	_level       _LEVEL
	_format      _FORMAT
	_rwLock      *sync.RWMutex
	_fileDir     string
	_fileName    string
	_maxSize     int64
	_unit        _UNIT
	_cutmode     _CUTMODE
	_mode        _MODE_TIME
	_filehandler *fileHandler
	_isFileWell  bool
	_formatter   string
	_maxBackup   int
	_isConsole   bool
	_gzip        bool
	_isTicker    int32
}

// return a new log object
func NewLogger() (log *Logging) {
	log = &Logging{_level: default_level, _cutmode: _TIMEMODE, _rwLock: new(sync.RWMutex), _format: default_format, _isConsole: true, _formatter: default_formatter}
	log.newfileHandler()
	return
}

// 控制台日志是否打开
func (t *Logging) SetConsole(_isConsole bool) *Logging {
	t._isConsole = _isConsole
	return t
}
func (t *Logging) Debug(v ...interface{}) *Logging {
	t.println(LEVEL_DEBUG, 2, v...)
	return t
}
func (t *Logging) Info(v ...interface{}) *Logging {
	t.println(LEVEL_INFO, 2, v...)
	return t
}
func (t *Logging) Warn(v ...interface{}) *Logging {
	t.println(LEVEL_WARN, 2, v...)
	return t
}
func (t *Logging) Error(v ...interface{}) *Logging {
	t.println(LEVEL_ERROR, 2, v...)
	return t
}
func (t *Logging) Fatal(v ...interface{}) *Logging {
	t.println(LEVEL_FATAL, 2, v...)
	return t
}

func (t *Logging) WriteBin(bs []byte) (bakfn string, err error) {
	if t._isFileWell {
		var openFileErr error
		if t._filehandler.mustBackUp() {
			bakfn, err, openFileErr = t.backUp()
		}
		if openFileErr == nil {
			t._rwLock.RLock()
			_, err = t._filehandler.write2file(bs)
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
		if t._filehandler.mustBackUp() {
			_, err, openFileErr = t.backUp()
		}
		if openFileErr == nil {
			t._rwLock.RLock()
			n, err = t._filehandler.write2file(bs)
			t._rwLock.RUnlock()
		}
	} else {
		err = errors.New("no log file found")
	}
	return
}

func (t *Logging) SetFormat(format _FORMAT) *Logging {
	t._format = format
	return t
}

func (t *Logging) SetLevel(level _LEVEL) *Logging {
	t._level = level
	return t
}

func (t *Logging) SetFormatter(formatter string) *Logging {
	t._formatter = formatter
	return t
}

/*
按日志文件大小分割日志文件
fileDir 日志文件夹路径
fileName 日志文件名
maxFileSize  日志文件大小最大值
unit    日志文件大小单位
*/
// Deprecated
// Use SeOption() instead.
func (t *Logging) SetRollingFile(fileDir, fileName string, maxFileSize int64, unit _UNIT) (l *Logging, err error) {
	return t.SetRollingFileLoop(fileDir, fileName, maxFileSize, unit, 0)
}

/*
按日志文件大小分割日志文件，指定保留的最大日志文件数
fileDir 日志文件夹路径
fileName 日志文件名
maxFileSize  日志文件大小最大值
unit    	日志文件大小单位
maxFileNum  留的日志文件数
*/
// Deprecated
// Use SeOption() instead.
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
	return t, err
}

/*
按日期分割日志文件
fileDir 日志文件夹路径
fileName 日志文件名
*/
// Deprecated
// Use SeOption() instead.
func (t *Logging) SetRollingDaily(fileDir, fileName string) (l *Logging, err error) {
	return t.SetRollingByTime(fileDir, fileName, MODE_DAY)
}

/*
指定按 小时，天，月 分割日志文件
fileDir 日志文件夹路径
fileName 日志文件名
mode   指定 小时，天，月
*/
// Deprecated
// Use SeOption() instead.
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
		go t.ticker(func() {
			defer catchError()
			if t._filehandler.mustBackUp() {
				t.backUp()
			}
		})
	}
	return t, err
}

// Deprecated
// Use SeOption() instead.
func (t *Logging) SetGzipOn(is bool) *Logging {
	t._gzip = is
	if t._filehandler != nil {
		t._filehandler._gzip = is
	}
	return t
}

// config object
func (t *Logging) SetOption(option *Option) *Logging {
	t._rwLock.Lock()
	defer t._rwLock.Unlock()
	if option.Format == 0 {
		option.Format = default_format
	}
	if option.Formatter == "" {
		option.Formatter = default_formatter
	}
	t._formatter = option.Formatter
	t._isConsole = option.Console
	t._format = option.Format

	t._level = option.Level
	if option.FileOption != nil {
		t._cutmode = option.FileOption.Cutmode()
		abspath, _ := filepath.Abs(option.FileOption.FilePath())
		dirPath := filepath.Dir(abspath)
		fileName := filepath.Base(option.FileOption.FilePath())
		if option.FileOption.Cutmode() == _SIZEMODE {
			if dirPath == "" {
				dirPath, _ = os.Getwd()
			}
			maxBackup := option.FileOption.MaxBuckup()
			maxsize := option.FileOption.MaxSize()
			t._cutmode = _SIZEMODE
			t._fileDir, t._fileName, t._maxSize, t._maxBackup, t._unit, t._gzip = dirPath, fileName, int64(maxsize), maxBackup, 1, option.FileOption.Compress()
			if t._maxSize <= 0 {
				t._maxSize = 1 << 30
			}
			if t._filehandler != nil {
				t._filehandler.close()
			}
			t.newfileHandler()
			if err := t._filehandler.openFileHandler(); err == nil {
				t._isFileWell = true
			}
		} else {
			if dirPath == "" {
				dirPath, _ = os.Getwd()
			}
			t._cutmode = _TIMEMODE
			t._fileDir, t._fileName, t._mode, t._cutmode, t._maxBackup, t._gzip = dirPath, fileName, option.FileOption.TimeMode(), _TIMEMODE, option.FileOption.MaxBuckup(), option.FileOption.Compress()
			if t._mode == 0 {
				t._mode = MODE_DAY
			}
			if t._filehandler != nil {
				t._filehandler.close()
			}
			t.newfileHandler()
			if err := t._filehandler.openFileHandler(); err == nil {
				t._isFileWell = true
				go t.ticker(func() {
					defer catchError()
					if t._filehandler.mustBackUp() {
						t.backUp()
					}
				})
			}
		}
	}
	return t
}

func (t *Logging) newfileHandler() {
	t._filehandler = new(fileHandler)
	t._filehandler._fileDir, t._filehandler._fileName, t._filehandler._maxSize, t._filehandler._cutmode, t._filehandler._unit, t._filehandler._maxbackup, t._filehandler._mode, t._filehandler._gzip = t._fileDir, t._fileName, t._maxSize, t._cutmode, t._unit, t._maxBackup, t._mode, t._gzip
}

func (t *Logging) backUp() (bakfn string, err, openFileErr error) {
	t._rwLock.Lock()
	defer t._rwLock.Unlock()
	if !t._filehandler.mustBackUp() {
		return
	}
	if err = t._filehandler.close(); err != nil {
		fprintln(t._format, LEVEL_ERROR, LEVEL_ERROR, 1, nil, err.Error())
		return
	}
	if bakfn, err = t._filehandler.rename(); err != nil {
		fprintln(t._format, LEVEL_ERROR, LEVEL_ERROR, 1, nil, err.Error())
		return
	}
	if openFileErr = t._filehandler.openFileHandler(); openFileErr != nil {
		fprintln(t._format, LEVEL_ERROR, LEVEL_ERROR, 1, nil, openFileErr.Error())
	}
	return
}

func (t *Logging) println(_level _LEVEL, calldepth int, v ...interface{}) {
	if t._level > _level {
		return
	}
	if t._isFileWell {
		var openFileErr error
		if t._filehandler.mustBackUp() {
			_, openFileErr, _ = t.backUp()
		}
		if openFileErr == nil {
			if t._format != FORMAT_NANO {
				bs := fmt.Append([]byte{}, v...)
				buf := getOutBuffer(bs, getlevelname(_level, t._format), t._format, k1(calldepth), &t._formatter)
				t._rwLock.RLock()
				t._filehandler.write2file(buf.Bytes())
				t._rwLock.RUnlock()
				buf.Free()
			} else {
				t._rwLock.RLock()
				t._filehandler.write2file(fmt.Appendln([]byte{}, v...))
				t._rwLock.RUnlock()
			}
		}
	}
	if t._isConsole {
		fprintln(t._format, _level, t._level, k1(calldepth), &t._formatter, v...)
	}
}

type fileHandler struct {
	_fileDir     string
	_fileName    string
	_maxSize     int64
	_fileSize    int64
	_unit        _UNIT
	_fileHandle  *os.File
	_cutmode     _CUTMODE
	_tomorSecond int64
	_maxbackup   int
	_mode        _MODE_TIME
	_gzip        bool
	_lastPrint   int64
}

func (t *fileHandler) openFileHandler() (e error) {
	if t._fileDir == "" || t._fileName == "" {
		e = errors.New("log filePath is null or error")
		return
	}
	e = mkdirDir(t._fileDir)
	if e != nil {
		return
	}
	fname := t._fileDir + "/" + t._fileName
	t._fileHandle, e = os.OpenFile(fname, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if e != nil {
		fprintln(default_format, LEVEL_ERROR, LEVEL_ERROR, 1, nil, e.Error())
		return
	}
	t._tomorSecond = tomorSecond(t._mode)
	if fs, err := t._fileHandle.Stat(); err == nil {
		t._fileSize = fs.Size()
	} else {
		e = err
	}
	return
}

func (t *fileHandler) addFileSize(size int64) {
	atomic.AddInt64(&t._fileSize, size)
}

func (t *fileHandler) write2file(bs []byte) (n int, e error) {
	defer catchError()
	if bs != nil {
		if n, e = _write2file(t._fileHandle, bs); e == nil {
			t.addFileSize(int64(n))
			if t._cutmode == _TIMEMODE {
				t._lastPrint = _time().Unix()
			}
		}
	}
	return
}

func (t *fileHandler) mustBackUp() bool {
	if t._fileSize == 0 {
		return false
	}
	switch t._cutmode {
	case _TIMEMODE:
		return _time().Unix() >= t._tomorSecond
	case _SIZEMODE:
		return t._fileSize > 0 && atomic.LoadInt64(&t._fileSize) >= t._maxSize*int64(t._unit)
	}
	return false
}

func (t *fileHandler) rename() (bckupfilename string, err error) {
	if t._cutmode == _TIMEMODE {
		bckupfilename = getBackupDayliFileName(t._lastPrint, t._fileDir, t._fileName, t._mode, t._gzip)
	} else {
		bckupfilename, err = getBackupRollFileName(t._fileDir, t._fileName, t._gzip)
	}
	if bckupfilename != "" && err == nil {
		oldPath := t._fileDir + "/" + t._fileName
		newPath := t._fileDir + "/" + bckupfilename
		if err = os.Rename(oldPath, newPath); err == nil {
			go func() {
				defer catchError()
				if t._gzip {
					if err = lgzip(newPath+".gz", bckupfilename, newPath); err == nil {
						os.Remove(newPath)
					}
				}
				if t._maxbackup > 0 {
					maxbuckup(t._fileDir, t._fileName, t._maxbackup)
				}
			}()
		}
	}
	return
}

func (t *fileHandler) close() (err error) {
	defer catchError()
	if t._fileHandle != nil {
		err = t._fileHandle.Close()
	}
	return
}

func tomorSecond(mode _MODE_TIME) int64 {
	now := _time()
	switch mode {
	case MODE_DAY:
		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()).Unix()
	case MODE_HOUR:
		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location()).Unix()
	case MODE_MONTH:
		return time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1).Unix()
	default:
		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()).Unix()
	}
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
//	//now := _time()
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
		if isFileExist(fmt.Sprint(dir, "/", bckupfilename, ".gz")) {
			bckupfilename = _getBackupfilename(1, dir, fmt.Sprint(fname, "_", timeStr), suffix, isGzip)
		}
	} else {
		if isFileExist(fmt.Sprint(dir, "/", bckupfilename)) {
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
		if _matchString(pattern, fd.Name()) {
			i++
		}
	}
	bckupfilename = _getBackupfilename(i, dir, fname, suffix, isGzip)
	return
}

func _getBackupfilename(count int, dir, filename, suffix string, isGzip bool) (bckupfilename string) {
	bckupfilename = fmt.Sprint(filename, "_", count, suffix)
	if isGzip {
		if isFileExist(fmt.Sprint(dir, "/", bckupfilename, ".gz")) {
			return _getBackupfilename(count+1, dir, filename, suffix, isGzip)
		}
	} else {
		if isFileExist(fmt.Sprint(dir, "/", bckupfilename)) {
			return _getBackupfilename(count+1, dir, filename, suffix, isGzip)
		}
	}
	return
}

func _write2file(f *os.File, bs []byte) (n int, e error) {
	n, e = f.Write(bs)
	return
}

func _console(s []byte, levelname []byte, flag _FORMAT, calldepth int, formatter *string) {
	if flag != FORMAT_NANO {
		buf := getOutBuffer(s, levelname, flag, k1(calldepth), formatter)
		fmt.Print(string(buf.Bytes()))
		buf.Free()
	} else {
		fmt.Println(string(s))
	}
}

func k1(calldepth int) int {
	return calldepth + 1
}

func getOutBuffer(s []byte, levelname []byte, format _FORMAT, calldepth int, formatter *string) *buffer.Buffer {
	return output(format, k1(calldepth), s, levelname, formatter)
}

func mkdirDir(dir string) (e error) {
	_, er := os.Stat(dir)
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(dir, 0777); err != nil {
			if os.IsPermission(err) {
				e = err
			}
		}
	}
	return
}

func isFileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func catchError() {
	if err := recover(); err != nil {
		fmt.Println(string(debug.Stack()))
	}
}

func maxbuckup(dir, filename string, maxcount int) {
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
					if _matchString(parrent, entry.Name()) {
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

func _matchString(pattern string, s string) bool {
	b, err := regexp.MatchString(pattern, s)
	if err != nil {
		b = false
	}
	return b
}

func _time() time.Time {
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

var m = hashmap.NewLimitMap[any, runtime.Frame](1 << 13)

func output(flag _FORMAT, calldepth int, s []byte, levelname []byte, formatter *string) (buf *buffer.Buffer) {
	now := _time()
	var file *string
	var line *int
	if flag&(FORMAT_SHORTFILENAME|FORMAT_LONGFILENAME) != 0 {
		var pcs [1]uintptr
		runtime.Callers(calldepth+1, pcs[:])
		var f runtime.Frame
		var ok bool
		if f, ok = m.Get(pcs); !ok {
			f, _ = runtime.CallersFrames([]uintptr{pcs[0]}).Next()
			m.Put(pcs, f)
		}
		file = &f.File
		line = &f.Line
	}
	return formatmsg(s, now, file, line, flag, levelname, formatter)
}

func formatmsg(msg []byte, t time.Time, file *string, line *int, flag _FORMAT, levelname []byte, formatter *string) (buf *buffer.Buffer) {
	buf = buffer.NewBufferByPool()
	var levelbuf *buffer.Buffer
	var timebuf *buffer.Buffer
	var filebuf *buffer.Buffer
	is_default_formatter := formatter == nil || (formatter != nil && (*formatter == default_formatter || *formatter == ""))
	if is_default_formatter {
		levelbuf, timebuf, filebuf = buf, buf, buf
	} else {
		levelbuf = buffer.NewBuffer()
		timebuf = buffer.NewBuffer()
		filebuf = buffer.NewBuffer()
	}
	if flag&FORMAT_LEVELFLAG != 0 {
		levelbuf.Write(levelname)
	}
	if flag&(FORMAT_DATE|FORMAT_TIME|FORMAT_MICROSECNDS) != 0 {
		if flag&FORMAT_DATE != 0 {
			year, month, day := t.Date()
			itoa(timebuf, year, 4)
			timebuf.WriteByte('/')
			itoa(timebuf, int(month), 2)
			timebuf.WriteByte('/')
			itoa(timebuf, day, 2)
			timebuf.WriteByte(' ')
		}
		if flag&(FORMAT_TIME|FORMAT_MICROSECNDS) != 0 {
			hour, min, sec := t.Clock()
			itoa(timebuf, hour, 2)
			timebuf.WriteByte(':')
			itoa(timebuf, min, 2)
			timebuf.WriteByte(':')
			itoa(timebuf, sec, 2)
			if flag&FORMAT_MICROSECNDS != 0 {
				timebuf.WriteByte('.')
				itoa(timebuf, t.Nanosecond()/1e3, 6)
			}
		}
		if is_default_formatter {
			timebuf.WriteByte(' ')
		}
	}
	if flag&(FORMAT_SHORTFILENAME|FORMAT_LONGFILENAME) != 0 {
		if flag&FORMAT_SHORTFILENAME != 0 {
			short := *file
			for i := len(*file) - 1; i > 0; i-- {
				if (*file)[i] == '/' {
					short = (*file)[i+1:]
					break
				}
			}
			file = &short
		}
		filebuf.Write([]byte(*file))
		filebuf.WriteByte(':')
		itoa(filebuf, *line, -1)

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

func itoa(buf *buffer.Buffer, i int, wid int) {
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
	buf.Write(b[bp:])
}

func (t *Logging) ticker(fn func()) {
	if !atomic.CompareAndSwapInt32(&t._isTicker, 0, 1) {
		return
	}
	for {
		waitTime := timeUntilNextWholeHour()
		if waitTime <= 0 {
			<-time.After(time.Second)
			continue
		}
		<-time.After(waitTime)
		fn()
	}
	atomic.CompareAndSwapInt32(&t._isTicker, 1, 0)
}

func timeUntilNextWholeHour() time.Duration {
	now := time.Now()
	nextWholeHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 1, now.Location())
	return nextWholeHour.Sub(now)
}
