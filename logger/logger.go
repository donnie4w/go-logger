package logger

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

/***
author:donnie email donnie4w@gmail.com
在控制台打印：直接调用 Debug(***) Info(***) Warn(***) Error(***) Fatal(***)
可以设置打印格式：SetFormat(FORMAT_SHORTFILENAME|FORMAT_DATE|FORMAT_TIME)
	无其他格式，只打印日志内容
	FORMAT_NANO
	长文件名及行数
	FORMAT_LONGFILENAME
	短文件名及行数
	FORMAT_SHORTFILENAME
	精确到日期
	FORMAT_DATE
	精确到秒
	FORMAT_TIME
	精确到微秒
	FORMAT_MICROSECNDS
	—————————————————————————————————————————————————————————————————————
    写日志文件可以获取实例
    全局实例可以直接调用log := logging.GetStaticLogger()
    获取新实例可以调用log := logging.NewLogger()
	1. 按日期分割日志文件
    	log.SetRollingDaily("d://foldTest", "log.txt")
	2. 按文件大小分割日志文件
	log.SetRollingFile("d://foldTest", "log.txt", 300, MB)
	log.SetConsole(false)控制台不打日志,默认值true
    日志级别
***/

const (
	_VER string = "1.0.1"
)

type _LEVEL int8
type _UNIT int64
type _TIME uint8
type _ROLLTYPE int //dailyRolling ,rollingFile
type _FORMAT int

const _DATEFORMAT = "20060102"

var static_mu *sync.Mutex = new(sync.Mutex)

// var _isGlobal = false
var static_lo *_logger = NewLogger()

const (
	_        = iota
	KB _UNIT = 1 << (iota * 10)
	MB
	GB
	TB
)

const (
	/*无其他格式，只打印日志内容*/
	FORMAT_NANO _FORMAT = 0
	/*长文件名及行数*/
	FORMAT_LONGFILENAME = _FORMAT(log.Llongfile)
	/*短文件名及行数*/
	FORMAT_SHORTFILENAME = _FORMAT(log.Lshortfile)
	/*日期时间精确到天*/
	FORMAT_DATE = _FORMAT(log.Ldate)
	/*时间精确到秒*/
	FORMAT_TIME = _FORMAT(log.Ltime)
	/*时间精确到微秒*/
	FORMAT_MICROSECNDS = _FORMAT(log.Lmicroseconds)
)

const (
	/*日志级别：ALL 最低级别*/
	ALL _LEVEL = iota
	/*日志级别：DEBUG 小于INFO*/
	DEBUG
	/*日志级别：INFO 小于 WARN*/
	INFO
	/*日志级别：WARN 小于 ERROR*/
	WARN
	/*日志级别：ERROR 小于 FATAL*/
	ERROR
	/*日志级别：FATAL 小于 OFF*/
	FATAL
	/*日志级别：off 不打印任何日志*/
	OFF
)

const (
	_DAILY _ROLLTYPE = iota
	_ROLLFILE
)

var default_format _FORMAT = FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME
var default_level = ALL

/*设置打印格式*/
func SetFormat(format _FORMAT) {
	default_format = format
	static_lo.SetFormat(format)

}

/*设置控制台日志级别，默认ALL*/
func SetLevel(level _LEVEL) {
	default_level = level
	static_lo.SetLevel(level)
}

/*获得全局Logger对象*/
func GetStaticLogger() *_logger {
	return _staticLogger()
}

func SetRollingFile(fileDir, fileName string, maxFileSize int64, unit _UNIT) (err error) {
	return static_lo.SetRollingFile(fileDir, fileName, maxFileSize, unit)
}

func SetRollingDaily(fileDir, fileName string) (err error) {
	return static_lo.SetRollingDaily(fileDir, fileName)
}

/**
设置全局log对象，默认false，则不获取对象时，默认控制台打印
true时，默认全局共用一个log对象，可设置 日志文件
*/
// func SetGlobal(isGlobal bool) *_logger {
// 	_isGlobal = isGlobal
// 	return _staticLogger()
// }

func _staticLogger() *_logger {
	// if static_lo != nil {
	// 	return static_lo
	// }
	// static_mu.Lock()
	// defer static_mu.Unlock()
	// if static_lo == nil {
	// 	static_lo = NewLogger()
	// }
	return static_lo
}

func Debug(v ...interface{}) {
	_print(default_format, DEBUG, default_level, 2, v...)
}
func Info(v ...interface{}) {
	_print(default_format, INFO, default_level, 2, v...)
}
func Warn(v ...interface{}) {
	_print(default_format, WARN, default_level, 2, v...)
}
func Error(v ...interface{}) {
	_print(default_format, ERROR, default_level, 2, v...)
}
func Fatal(v ...interface{}) {
	_print(default_format, FATAL, default_level, 2, v...)
}

func _print(_format _FORMAT, level, _default_level _LEVEL, calldepth int, v ...interface{}) {
	if level < _default_level {
		return
	}
	// if _isGlobal {
	_staticLogger().println(level, k1(calldepth), v...)
	// } else {
	// 	_console(fmt.Sprint(v...), getlevelname(level, default_format), _format, k1(calldepth))
	// }
}

func __print(_format _FORMAT, level, _default_level _LEVEL, calldepth int, v ...interface{}) {
	_console(fmt.Sprint(v...), getlevelname(level, default_format), _format, k1(calldepth))
}

func getlevelname(level _LEVEL, format _FORMAT) (levelname string) {
	if format == FORMAT_NANO {
		return
	}
	switch level {
	case ALL:
		levelname = "[ALL]"
	case DEBUG:
		levelname = "[DEBUG]"
	case INFO:
		levelname = "[INFO]"
	case WARN:
		levelname = "[WARN]"
	case ERROR:
		levelname = "[ERROR]"
	case FATAL:
		levelname = "[FATAL]"
	default:
	}
	return
}

/*————————————————————————————————————————————————————————————————————————————*/
type _logger struct {
	_level     _LEVEL
	_format    _FORMAT
	_rwLock    *sync.RWMutex
	_safe      bool
	_fileDir   string
	_fileName  string
	_maxSize   int64
	_unit      _UNIT
	_rolltype  _ROLLTYPE
	_fileObj   *fileObj
	_isConsole bool
}

func NewLogger() (log *_logger) {
	log = &_logger{_level: DEBUG, _rolltype: _DAILY, _rwLock: new(sync.RWMutex), _format: FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME, _isConsole: true}
	log.newfileObj()
	return
}

//控制台日志是否打开
func (this *_logger) SetConsole(_isConsole bool) {
	this._isConsole = _isConsole
}
func (this *_logger) Debug(v ...interface{}) {
	this.println(DEBUG, 2, v...)
}
func (this *_logger) Info(v ...interface{}) {
	this.println(INFO, 2, v...)
}
func (this *_logger) Warn(v ...interface{}) {
	this.println(WARN, 2, v...)
}
func (this *_logger) Error(v ...interface{}) {
	this.println(ERROR, 2, v...)
}
func (this *_logger) Fatal(v ...interface{}) {
	this.println(FATAL, 2, v...)
}
func (this *_logger) SetFormat(format _FORMAT) {
	this._format = format
}
func (this *_logger) SetLevel(level _LEVEL) {
	this._level = level
}

/*按日志文件大小分割日志文件
fileDir 日志文件夹路径
fileName 日志文件名
maxFileSize  日志文件大小最大值
unit    日志文件大小单位
*/
func (this *_logger) SetRollingFile(fileDir, fileName string, maxFileSize int64, unit _UNIT) (err error) {
	if fileDir == "" {
		fileDir, _ = os.Getwd()
	}
	this._fileDir, this._fileName, this._maxSize, this._unit = fileDir, fileName, maxFileSize, unit
	this._rolltype = _ROLLFILE
	if this._fileObj != nil {
		this._fileObj.close()
	}
	this.newfileObj()
	err = this._fileObj.openFileHandler()
	if err != nil {
		Fatal(err.Error())
	}
	return
}

/*按日期分割日志文件
fileDir 日志文件夹路径
fileName 日志文件名
*/
func (this *_logger) SetRollingDaily(fileDir, fileName string) (err error) {
	if fileDir == "" {
		fileDir, _ = os.Getwd()
	}
	this._fileDir, this._fileName = fileDir, fileName
	this._rolltype = _DAILY
	if this._fileObj != nil {
		this._fileObj.close()
	}
	this.newfileObj()
	err = this._fileObj.openFileHandler()
	if err != nil {
		Fatal(err.Error())
	}
	return
}

func (this *_logger) SetRollingByTime(fileDir, fileName string, rolltype _TIME) (err error) {
	return
}

func (this *_logger) newfileObj() {
	this._fileObj = new(fileObj)
	this._fileObj._fileDir, this._fileObj._fileName, this._fileObj._maxSize, this._fileObj._rolltype, this._fileObj._unit = this._fileDir, this._fileName, this._maxSize, this._rolltype, this._unit
}

func (this *_logger) backUp() {
	this._rwLock.Lock()
	defer this._rwLock.Unlock()
	if !this._fileObj.isMustBackUp() {
		return
	}
	this._fileObj.close()
	err := this._fileObj.rename()
	if err != nil {
		_print(this._format, FATAL, FATAL, 4, err.Error())
	}
	this._fileObj.openFileHandler()
}

func (this *_logger) println(_level _LEVEL, calldepth int, v ...interface{}) {
	if this._level > _level {
		return
	}
	if this._fileObj._isFileWell {
		if this._fileObj.isMustBackUp() {
			this.backUp()
		}
		func() {
			this._rwLock.RLock()
			defer this._rwLock.RUnlock()
			s := fmt.Sprint(v...)
			buf := getOutBuffer(s, getlevelname(_level, this._format), this._format, k1(calldepth)+1)
			this._fileObj.write2file(buf.Bytes())
		}()
	}
	if this._isConsole {
		__print(this._format, _level, this._level, k1(calldepth), v...)
	}
}

/*————————————————————————————————————————————————————————————————————————————*/
type fileObj struct {
	_fileDir     string
	_fileName    string
	_maxSize     int64
	_fileSize    int64
	_unit        _UNIT
	_fileHandler *os.File
	_rolltype    _ROLLTYPE
	// _fileCreateDate *time.Time
	_tomorSecond int64
	_isFileWell  bool
}

func (this *fileObj) openFileHandler() (e error) {
	if this._fileDir == "" || this._fileName == "" {
		e = errors.New("log filePath is null or error")
		return
	}
	e = mkdirDir(this._fileDir)
	if e != nil {
		this._isFileWell = false
		return
	}
	fname := fmt.Sprint(this._fileDir, "/", this._fileName)
	this._fileHandler, e = os.OpenFile(fname, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if e != nil {
		_print(default_format, FATAL, FATAL, 4, e.Error())
		this._isFileWell = false
		return
	}
	this._isFileWell = true
	this._tomorSecond = tomorSecond()
	fs, err := this._fileHandler.Stat()
	if err == nil {
		this._fileSize = fs.Size()
	} else {
		e = err
	}
	return
}

func (this *fileObj) addFileSize(size int64) {
	atomic.AddInt64(&this._fileSize, size)
}

func (this *fileObj) write2file(bs []byte) (e error) {
	defer catchError()
	if bs != nil {
		this.addFileSize(int64(len(bs)))
		_write2file(this._fileHandler, bs)
	}
	return
}

func (this *fileObj) isMustBackUp() bool {
	switch this._rolltype {
	case _DAILY:
		if time.Now().Unix() >= this._tomorSecond {
			return true
		}
	case _ROLLFILE:
		return this._fileSize > 0 && this._fileSize >= this._maxSize*int64(this._unit)
	}
	return false
}

func (this *fileObj) rename() (err error) {
	bckupfilename := ""
	if this._rolltype == _DAILY {
		bckupfilename = getBackupDayliFileName(this._fileDir, this._fileName)
	} else {
		bckupfilename, err = getBackupRollFileName(this._fileDir, this._fileName)
	}
	if bckupfilename != "" && err == nil {
		oldPath := fmt.Sprint(this._fileDir, "/", this._fileName)
		newPath := fmt.Sprint(this._fileDir, "/", bckupfilename)
		err = os.Rename(oldPath, newPath)
	}
	return
}

func (this *fileObj) close() {
	defer catchError()
	if this._fileHandler != nil {
		this._fileHandler.Close()
	}
}

func tomorSecond() int64 {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()).Unix()
}

/*————————————————————————————————————————————————————————————————————————————*/
func getBackupDayliFileName(dir, filename string) (bckupfilename string) {
	timeStr := time.Now().AddDate(0, 0, -1).Format(_DATEFORMAT)
	index := strings.LastIndex(filename, ".")
	if index <= 0 {
		index = len(filename)
	}
	fname := filename[:index]
	suffix := filename[index:]
	bckupfilename = fmt.Sprint(fname, "_", timeStr, suffix)
	if isFileExist(fmt.Sprint(dir, "/", bckupfilename)) {
		bckupfilename = _getBackupfilename(1, dir, fmt.Sprint(fname, "_", timeStr), suffix)
	}
	return
}

func getBackupRollFileName(dir, filename string) (bckupfilename string, er error) {
	list, err := os.ReadDir(dir)
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
	length := len(list)
	bckupfilename = _getBackupfilename(length, dir, fname, suffix)
	return
}

func _getBackupfilename(count int, dir, filename, suffix string) (bckupfilename string) {
	bckupfilename = fmt.Sprint(filename, "_", count, suffix)
	if isFileExist(fmt.Sprint(dir, "/", bckupfilename)) {
		return _getBackupfilename(count+1, dir, filename, suffix)
	}
	return
}

func _write2file(f *os.File, bs []byte) (e error) {
	_, e = f.Write(bs)
	return
}

func _console(s string, levelname string, flag _FORMAT, calldepth int) {
	buf := getOutBuffer(s, levelname, flag, k1(calldepth))
	fmt.Print(&buf)
}

func outwriter(out io.Writer, prefix string, flag _FORMAT, calldepth int, s string) {
	l := log.New(out, prefix, int(flag))
	l.Output(k1(calldepth), s)
}

func k1(calldepth int) int {
	return calldepth + 1
}

func getOutBuffer(s string, levelname string, flag _FORMAT, calldepth int) (buf bytes.Buffer) {
	outwriter(&buf, levelname, flag, k1(calldepth), s)
	return
}

func mkdirDir(dir string) (e error) {
	_, er := os.Stat(dir)
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(dir, 0666); err != nil {
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
		Fatal(string(debug.Stack()))
	}
}
