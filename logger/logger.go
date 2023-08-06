// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.

package logger

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	bufpool "github.com/donnie4w/gofer/pool/buffer"
)

const (
	_VER string = "2.0.2"
)

type _LEVEL int8
type _UNIT int64
type _MODE_TIME uint8
type _ROLLTYPE int //dailyRolling ,rollingFile
type _FORMAT int

const _DATEFORMAT_DAY = "20060102"
const _DATEFORMAT_HOUR = "2006010215"
const _DATEFORMAT_MONTH = "200601"

var static_mu *sync.Mutex = new(sync.Mutex)

var static_lo *Logging = NewLogger()

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
	FORMAT_NANO       _FORMAT = 0

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
	_DAYLY _ROLLTYPE = iota
	_ROLLFILE
)

var default_format _FORMAT = FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME
var default_level = LEVEL_ALL

/*设置打印格式*/
func SetFormat(format _FORMAT) *Logging {
	default_format = format
	return static_lo.SetFormat(format)
}

/*设置控制台日志级别，默认ALL*/
// Setting the log Level
func SetLevel(level _LEVEL) *Logging {
	default_level = level
	return static_lo.SetLevel(level)
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
func SetRollingFile(fileDir, fileName string, maxFileSize int64, unit _UNIT) (l *Logging, err error) {
	return SetRollingFileLoop(fileDir, fileName, maxFileSize, unit, 0)
}

// yesterday's log data is backed up to a specified log file each day
// Parameters:
//   - fileDir   :directory where log files are stored, If it is the current directory, you also can set it to ""
//   - fileName  : log file name
func SetRollingDaily(fileDir, fileName string) (l *Logging, err error) {
	return SetRollingByTime(fileDir, fileName, MODE_DAY)
}

// like SetRollingFile,but only keep (maxFileNum) current files
// - maxFileNum : the number of files that are retained
func SetRollingFileLoop(fileDir, fileName string, maxFileSize int64, unit _UNIT, maxFileNum int) (l *Logging, err error) {
	return static_lo.SetRollingFileLoop(fileDir, fileName, maxFileSize, unit, maxFileNum)
}

// like SetRollingDaily,but supporte hourly backup ,dayly backup and monthly backup
// mode : 	MODE_HOUR    MODE_DAY   MODE_MONTH
func SetRollingByTime(fileDir, fileName string, mode _MODE_TIME) (l *Logging, err error) {
	return static_lo.SetRollingByTime(fileDir, fileName, mode)
}

// when set true, the specified backup file of both SetRollingFile and SetRollingFileLoop will be save as a compressed file
func SetGzipOn(is bool) (l *Logging) {
	return static_lo.SetGzipOn(is)
}

func _staticLogger() *Logging {
	return static_lo
}

// Logs are printed at the DEBUG level
func Debug(v ...interface{}) *Logging {
	_print(default_format, LEVEL_DEBUG, default_level, 2, v...)
	return _staticLogger()
}

// Logs are printed at the INFO level
func Info(v ...interface{}) *Logging {
	_print(default_format, LEVEL_INFO, default_level, 2, v...)
	return _staticLogger()
}

// Logs are printed at the WARN level
func Warn(v ...interface{}) *Logging {
	_print(default_format, LEVEL_WARN, default_level, 2, v...)
	return _staticLogger()
}

// Logs are printed at the ERROR level
func Error(v ...interface{}) *Logging {
	_print(default_format, LEVEL_ERROR, default_level, 2, v...)
	return _staticLogger()
}

// Logs are printed at the FATAL level
func Fatal(v ...interface{}) *Logging {
	_print(default_format, LEVEL_FATAL, default_level, 2, v...)
	return _staticLogger()
}

func _print(_format _FORMAT, level, _default_level _LEVEL, calldepth int, v ...interface{}) {
	if level < _default_level {
		return
	}
	_staticLogger().println(level, k1(calldepth), v...)
}

func __print(_format _FORMAT, level, _default_level _LEVEL, calldepth int, v ...interface{}) {
	_console(fmt.Append([]byte{}, v...), getlevelname(level, default_format), _format, k1(calldepth))
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
	_level      _LEVEL
	_format     _FORMAT
	_rwLock     *sync.RWMutex
	_safe       bool
	_fileDir    string
	_fileName   string
	_maxSize    int64
	_unit       _UNIT
	_rolltype   _ROLLTYPE
	_mode       _MODE_TIME
	_fileObj    *fileObj
	_maxFileNum int
	_isConsole  bool
	_gzip       bool
}

// return a new log object
func NewLogger() (log *Logging) {
	log = &Logging{_level: LEVEL_DEBUG, _rolltype: _DAYLY, _rwLock: new(sync.RWMutex), _format: FORMAT_SHORTFILENAME | FORMAT_DATE | FORMAT_TIME, _isConsole: true}
	log.newfileObj()
	return
}

// 控制台日志是否打开
func (this *Logging) SetConsole(_isConsole bool) *Logging {
	this._isConsole = _isConsole
	return this
}
func (this *Logging) Debug(v ...interface{}) *Logging {
	this.println(LEVEL_DEBUG, 2, v...)
	return this
}
func (this *Logging) Info(v ...interface{}) *Logging {
	this.println(LEVEL_INFO, 2, v...)
	return this
}
func (this *Logging) Warn(v ...interface{}) *Logging {
	this.println(LEVEL_WARN, 2, v...)
	return this
}
func (this *Logging) Error(v ...interface{}) *Logging {
	this.println(LEVEL_ERROR, 2, v...)
	return this
}
func (this *Logging) Fatal(v ...interface{}) *Logging {
	this.println(LEVEL_FATAL, 2, v...)
	return this
}

func (this *Logging) Write(bs []byte) (err error, bakfn string) {
	if this._fileObj._isFileWell {
		var openFileErr error
		if this._fileObj.isMustBackUp() {
			err, openFileErr, bakfn = this.backUp()
		}
		if openFileErr == nil {
			this._rwLock.RLock()
			defer this._rwLock.RUnlock()
			_, err = this._fileObj.write2file(bs)
			return
		}
	}
	return
}

func (this *Logging) SetFormat(format _FORMAT) *Logging {
	this._format = format
	return this
}
func (this *Logging) SetLevel(level _LEVEL) *Logging {
	this._level = level
	return this
}

/*
按日志文件大小分割日志文件
fileDir 日志文件夹路径
fileName 日志文件名
maxFileSize  日志文件大小最大值
unit    日志文件大小单位
*/
func (this *Logging) SetRollingFile(fileDir, fileName string, maxFileSize int64, unit _UNIT) (l *Logging, err error) {
	return this.SetRollingFileLoop(fileDir, fileName, maxFileSize, unit, 0)
}

/*
按日志文件大小分割日志文件，指定保留的最大日志文件数
fileDir 日志文件夹路径
fileName 日志文件名
maxFileSize  日志文件大小最大值
unit    	日志文件大小单位
maxFileNum  留的日志文件数
*/
func (this *Logging) SetRollingFileLoop(fileDir, fileName string, maxFileSize int64, unit _UNIT, maxFileNum int) (l *Logging, err error) {
	if fileDir == "" {
		fileDir, _ = os.Getwd()
	}
	if maxFileNum > 0 {
		maxFileNum--
	}
	this._fileDir, this._fileName, this._maxSize, this._maxFileNum, this._unit = fileDir, fileName, maxFileSize, maxFileNum, unit
	this._rolltype = _ROLLFILE
	if this._fileObj != nil {
		this._fileObj.close()
	}
	this.newfileObj()
	err = this._fileObj.openFileHandler()
	return this, err
}

/*
按日期分割日志文件
fileDir 日志文件夹路径
fileName 日志文件名
*/
func (this *Logging) SetRollingDaily(fileDir, fileName string) (l *Logging, err error) {
	return this.SetRollingByTime(fileDir, fileName, MODE_DAY)
}

/*
指定按 小时，天，月 分割日志文件
fileDir 日志文件夹路径
fileName 日志文件名
mode   指定 小时，天，月
*/
func (this *Logging) SetRollingByTime(fileDir, fileName string, mode _MODE_TIME) (l *Logging, err error) {
	if fileDir == "" {
		fileDir, _ = os.Getwd()
	}
	this._fileDir, this._fileName, this._mode = fileDir, fileName, mode
	this._rolltype = _DAYLY
	if this._fileObj != nil {
		this._fileObj.close()
	}
	this.newfileObj()
	err = this._fileObj.openFileHandler()
	return this, err
}

func (this *Logging) SetGzipOn(is bool) *Logging {
	this._gzip = is
	if this._fileObj != nil {
		this._fileObj._gzip = is
	}
	return this
}

func (this *Logging) newfileObj() {
	this._fileObj = new(fileObj)
	this._fileObj._fileDir, this._fileObj._fileName, this._fileObj._maxSize, this._fileObj._rolltype, this._fileObj._unit, this._fileObj._maxFileNum, this._fileObj._mode, this._fileObj._gzip = this._fileDir, this._fileName, this._maxSize, this._rolltype, this._unit, this._maxFileNum, this._mode, this._gzip
}

func (this *Logging) backUp() (err, openFileErr error, bakfn string) {
	this._rwLock.Lock()
	defer this._rwLock.Unlock()
	if !this._fileObj.isMustBackUp() {
		return
	}
	err = this._fileObj.close()
	if err != nil {
		__print(this._format, LEVEL_ERROR, LEVEL_ERROR, 1, err.Error())
		return
	}
	err, bakfn = this._fileObj.rename()
	if err != nil {
		__print(this._format, LEVEL_ERROR, LEVEL_ERROR, 1, err.Error())
		return
	}
	openFileErr = this._fileObj.openFileHandler()
	if openFileErr != nil {
		__print(this._format, LEVEL_ERROR, LEVEL_ERROR, 1, openFileErr.Error())
	}
	return
}

func (this *Logging) println(_level _LEVEL, calldepth int, v ...interface{}) {
	if this._level > _level {
		return
	}
	if this._fileObj._isFileWell {
		var openFileErr error
		if this._fileObj.isMustBackUp() {
			_, openFileErr, _ = this.backUp()
		}
		if openFileErr == nil {
			func() {
				this._rwLock.RLock()
				defer this._rwLock.RUnlock()
				if this._format != FORMAT_NANO {
					bs := fmt.Append([]byte{}, v...)
					buf := getOutBuffer(bs, getlevelname(_level, this._format), this._format, k1(calldepth)+1)
					this._fileObj.write2file(buf.Bytes())
					bufferpool.Put(buf)
				} else {
					bs := make([]byte, 0)
					this._fileObj.write2file(fmt.Appendln(bs, v...))
				}
			}()
		}
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
	_tomorSecond int64
	_isFileWell  bool
	_maxFileNum  int
	_mode        _MODE_TIME
	_gzip        bool
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
		__print(default_format, LEVEL_ERROR, LEVEL_ERROR, 1, e.Error())
		this._isFileWell = false
		return
	}
	this._isFileWell = true
	this._tomorSecond = tomorSecond(this._mode)
	if fs, err := this._fileHandler.Stat(); err == nil {
		this._fileSize = fs.Size()
	} else {
		e = err
	}
	return
}

func (this *fileObj) addFileSize(size int64) {
	atomic.AddInt64(&this._fileSize, size)
}

func (this *fileObj) write2file(bs []byte) (n int, e error) {
	defer catchError()
	if bs != nil {
		if n, e = _write2file(this._fileHandler, bs); e == nil {
			this.addFileSize(int64(n))
		}
	}
	return
}

func (this *fileObj) isMustBackUp() bool {
	switch this._rolltype {
	case _DAYLY:
		if _time().Unix() >= this._tomorSecond {
			return true
		}
	case _ROLLFILE:
		return this._fileSize > 0 && this._fileSize >= this._maxSize*int64(this._unit)
	}
	return false
}

func (this *fileObj) rename() (err error, bckupfilename string) {
	if this._rolltype == _DAYLY {
		bckupfilename = getBackupDayliFileName(this._fileDir, this._fileName, this._mode, this._gzip)
	} else {
		bckupfilename, err = getBackupRollFileName(this._fileDir, this._fileName, this._gzip)
	}
	if bckupfilename != "" && err == nil {
		oldPath := fmt.Sprint(this._fileDir, "/", this._fileName)
		newPath := fmt.Sprint(this._fileDir, "/", bckupfilename)
		err = os.Rename(oldPath, newPath)
		go func() {
			if err == nil && this._gzip {
				if err = lgzip(fmt.Sprint(newPath, ".gz"), bckupfilename, newPath); err == nil {
					os.Remove(newPath)
				}
			}
			if err == nil && this._rolltype == _ROLLFILE && this._maxFileNum > 0 {
				_rmOverCountFile(this._fileDir, bckupfilename, this._maxFileNum, this._gzip)
			}
		}()
	}
	return
}

func (this *fileObj) close() (err error) {
	defer catchError()
	if this._fileHandler != nil {
		err = this._fileHandler.Close()
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

func _yestStr(mode _MODE_TIME) string {
	now := _time()
	switch mode {
	case MODE_DAY:
		return now.AddDate(0, 0, -1).Format(_DATEFORMAT_DAY)
	case MODE_HOUR:
		return now.Add(-1 * time.Hour).Format(_DATEFORMAT_HOUR)
	case MODE_MONTH:
		return now.AddDate(0, -1, 0).Format(_DATEFORMAT_MONTH)
	default:
		return now.AddDate(0, 0, -1).Format(_DATEFORMAT_DAY)
	}
}

var bufferpool = bufpool.NewBufferPool(6)

func getBackupDayliFileName(dir, filename string, mode _MODE_TIME, isGzip bool) (bckupfilename string) {
	timeStr := _yestStr(mode)
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

func _console(s []byte, levelname []byte, flag _FORMAT, calldepth int) {
	if flag != FORMAT_NANO {
		buf := getOutBuffer(s, levelname, flag, k1(calldepth))
		fmt.Print(string(buf.Bytes()))
		bufferpool.Put(buf)
	} else {
		fmt.Println(string(s))
	}
}

func k1(calldepth int) int {
	return calldepth + 1
}

func getOutBuffer(s []byte, levelname []byte, flag _FORMAT, calldepth int) (buf *bytes.Buffer) {
	buf = output(flag, k1(calldepth), s, levelname)
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

func _rmOverCountFile(dir, backupfileName string, maxFileNum int, isGzip bool) {
	static_mu.Lock()
	defer static_mu.Unlock()
	f, err := os.Open(dir)
	if err != nil {
		return
	}
	dirs, _ := f.ReadDir(-1)
	f.Close()
	if len(dirs) <= maxFileNum {
		return
	}
	sort.Slice(dirs, func(i, j int) bool {
		f1, _ := dirs[i].Info()
		f2, _ := dirs[j].Info()
		return f1.ModTime().Unix() > f2.ModTime().Unix()
	})
	index := strings.LastIndex(backupfileName, "_")
	indexSuffix := strings.LastIndex(backupfileName, ".")
	if indexSuffix == 0 {
		indexSuffix = len(backupfileName)
	}
	prefixname := backupfileName[:index+1]
	suffix := backupfileName[indexSuffix:]
	suffixlen := len(suffix)
	rmfiles := make([]string, 0)
	i := 0
	for _, f := range dirs {
		checkfname := f.Name()
		if isGzip && strings.HasSuffix(checkfname, ".gz") {
			checkfname = checkfname[:len(checkfname)-3]
		}
		if len(checkfname) > len(prefixname) && checkfname[:len(prefixname)] == prefixname && _matchString("^[0-9]+$", checkfname[len(prefixname):len(checkfname)-suffixlen]) {
			finfo, err := f.Info()
			if err == nil && !finfo.IsDir() {
				i++
				if i > maxFileNum {
					rmfiles = append(rmfiles, fmt.Sprint(dir, "/", f.Name()))
				}
			}
		}
	}
	if len(rmfiles) > 0 {
		for _, k := range rmfiles {
			os.Remove(k)
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

func output(flag _FORMAT, calldepth int, s []byte, levelname []byte) (buf *bytes.Buffer) {
	now := _time()
	var file string
	var line int
	if flag&(FORMAT_SHORTFILENAME|FORMAT_LONGFILENAME) != 0 {
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
	}
	buf = bufferpool.Get(len(s) + formatHeaderLength(now, file, line, flag, levelname))
	formatHeader(buf, now, file, line, flag, levelname)
	buf.Write(s)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		buf.WriteByte('\n')
	}
	return
}

func formatHeader(buf *bytes.Buffer, t time.Time, file string, line int, flag _FORMAT, levelname []byte) {
	buf.Write(levelname)
	if flag&(FORMAT_DATE|FORMAT_TIME|FORMAT_MICROSECNDS) != 0 {
		if flag&FORMAT_DATE != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			buf.WriteByte('/')
			itoa(buf, int(month), 2)
			buf.WriteByte('/')
			itoa(buf, day, 2)
			buf.WriteByte(' ')
		}
		if flag&(FORMAT_TIME|FORMAT_MICROSECNDS) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			buf.WriteByte(':')
			itoa(buf, min, 2)
			buf.WriteByte(':')
			itoa(buf, sec, 2)
			if flag&FORMAT_MICROSECNDS != 0 {
				buf.WriteByte('.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			buf.WriteByte(' ')
		}
	}
	if flag&(FORMAT_SHORTFILENAME|FORMAT_LONGFILENAME) != 0 {
		if flag&FORMAT_SHORTFILENAME != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		buf.Write([]byte(file))
		buf.WriteByte(':')
		itoa(buf, line, -1)
		buf.WriteByte(':')
		buf.WriteByte(' ')
	}
}

func formatHeaderLength(t time.Time, file string, line int, flag _FORMAT, levelname []byte) (i int) {
	i += len(levelname)
	if flag&(FORMAT_DATE|FORMAT_TIME|FORMAT_MICROSECNDS) != 0 {
		if flag&FORMAT_DATE != 0 {
			i += 11
		}
		if flag&(FORMAT_TIME|FORMAT_MICROSECNDS) != 0 {
			i += 8
			if flag&FORMAT_MICROSECNDS != 0 {
				i += 7
			}
			i += 1
		}
	}
	if flag&(FORMAT_SHORTFILENAME|FORMAT_LONGFILENAME) != 0 {
		if flag&FORMAT_SHORTFILENAME != 0 {
			for k := len(file) - 1; k > 0; k-- {
				if file[k] == '/' {
					i += len(file) - k
					break
				}
			}
		} else {
			i += len(file)
		}
		i += 4
	}
	return
}

func itoa(buf *bytes.Buffer, i int, wid int) {
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
