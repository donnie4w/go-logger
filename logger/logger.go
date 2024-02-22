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

	. "github.com/donnie4w/gofer/buffer"
	. "github.com/donnie4w/gofer/hashmap"
)

const (
	_VER string = "0.23.0"
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

func (t *Logging) Write(bs []byte) (bakfn string, err error) {
	if t._fileObj._isFileWell {
		var openFileErr error
		if t._fileObj.isMustBackUp() {
			bakfn, err, openFileErr = t.backUp()
		}
		if openFileErr == nil {
			t._rwLock.RLock()
			defer t._rwLock.RUnlock()
			_, err = t._fileObj.write2file(bs)
			return
		}
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

/*
按日志文件大小分割日志文件
fileDir 日志文件夹路径
fileName 日志文件名
maxFileSize  日志文件大小最大值
unit    日志文件大小单位
*/
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
func (t *Logging) SetRollingFileLoop(fileDir, fileName string, maxFileSize int64, unit _UNIT, maxFileNum int) (l *Logging, err error) {
	if fileDir == "" {
		fileDir, _ = os.Getwd()
	}
	if maxFileNum > 0 {
		maxFileNum--
	}
	t._fileDir, t._fileName, t._maxSize, t._maxFileNum, t._unit = fileDir, fileName, maxFileSize, maxFileNum, unit
	t._rolltype = _ROLLFILE
	if t._fileObj != nil {
		t._fileObj.close()
	}
	t.newfileObj()
	err = t._fileObj.openFileHandler()
	return t, err
}

/*
按日期分割日志文件
fileDir 日志文件夹路径
fileName 日志文件名
*/
func (t *Logging) SetRollingDaily(fileDir, fileName string) (l *Logging, err error) {
	return t.SetRollingByTime(fileDir, fileName, MODE_DAY)
}

/*
指定按 小时，天，月 分割日志文件
fileDir 日志文件夹路径
fileName 日志文件名
mode   指定 小时，天，月
*/
func (t *Logging) SetRollingByTime(fileDir, fileName string, mode _MODE_TIME) (l *Logging, err error) {
	if fileDir == "" {
		fileDir, _ = os.Getwd()
	}
	t._fileDir, t._fileName, t._mode = fileDir, fileName, mode
	t._rolltype = _DAYLY
	if t._fileObj != nil {
		t._fileObj.close()
	}
	t.newfileObj()
	err = t._fileObj.openFileHandler()
	return t, err
}

func (t *Logging) SetGzipOn(is bool) *Logging {
	t._gzip = is
	if t._fileObj != nil {
		t._fileObj._gzip = is
	}
	return t
}

func (t *Logging) newfileObj() {
	t._fileObj = new(fileObj)
	t._fileObj._fileDir, t._fileObj._fileName, t._fileObj._maxSize, t._fileObj._rolltype, t._fileObj._unit, t._fileObj._maxFileNum, t._fileObj._mode, t._fileObj._gzip = t._fileDir, t._fileName, t._maxSize, t._rolltype, t._unit, t._maxFileNum, t._mode, t._gzip
}

func (t *Logging) backUp() (bakfn string, err, openFileErr error) {
	t._rwLock.Lock()
	defer t._rwLock.Unlock()
	if !t._fileObj.isMustBackUp() {
		return
	}
	err = t._fileObj.close()
	if err != nil {
		__print(t._format, LEVEL_ERROR, LEVEL_ERROR, 1, err.Error())
		return
	}
	bakfn, err = t._fileObj.rename()
	if err != nil {
		__print(t._format, LEVEL_ERROR, LEVEL_ERROR, 1, err.Error())
		return
	}
	openFileErr = t._fileObj.openFileHandler()
	if openFileErr != nil {
		__print(t._format, LEVEL_ERROR, LEVEL_ERROR, 1, openFileErr.Error())
	}
	return
}

func (t *Logging) println(_level _LEVEL, calldepth int, v ...interface{}) {
	if t._level > _level {
		return
	}
	if t._fileObj._isFileWell {
		var openFileErr error
		if t._fileObj.isMustBackUp() {
			_, openFileErr, _ = t.backUp()
		}
		if openFileErr == nil {
			func() {
				if t._format != FORMAT_NANO {
					bs := fmt.Append([]byte{}, v...)
					buf := getOutBuffer(bs, getlevelname(_level, t._format), t._format, k1(calldepth)+1)
					t._rwLock.RLock()
					defer t._rwLock.RUnlock()
					t._fileObj.write2file(buf.Bytes())
					buf.Free()
				} else {
					bs := make([]byte, 0)
					t._fileObj.write2file(fmt.Appendln(bs, v...))
				}
			}()
		}
	}
	if t._isConsole {
		__print(t._format, _level, t._level, k1(calldepth), v...)
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

func (t *fileObj) openFileHandler() (e error) {
	if t._fileDir == "" || t._fileName == "" {
		e = errors.New("log filePath is null or error")
		return
	}
	e = mkdirDir(t._fileDir)
	if e != nil {
		t._isFileWell = false
		return
	}
	fname := fmt.Sprint(t._fileDir, "/", t._fileName)
	t._fileHandler, e = os.OpenFile(fname, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if e != nil {
		__print(default_format, LEVEL_ERROR, LEVEL_ERROR, 1, e.Error())
		t._isFileWell = false
		return
	}
	t._isFileWell = true
	t._tomorSecond = tomorSecond(t._mode)
	if fs, err := t._fileHandler.Stat(); err == nil {
		t._fileSize = fs.Size()
	} else {
		e = err
	}
	return
}

func (t *fileObj) addFileSize(size int64) {
	atomic.AddInt64(&t._fileSize, size)
}

func (t *fileObj) write2file(bs []byte) (n int, e error) {
	defer catchError()
	if bs != nil {
		if n, e = _write2file(t._fileHandler, bs); e == nil {
			t.addFileSize(int64(n))
		}
	}
	return
}

func (t *fileObj) isMustBackUp() bool {
	switch t._rolltype {
	case _DAYLY:
		if _time().Unix() >= t._tomorSecond {
			return true
		}
	case _ROLLFILE:
		return t._fileSize > 0 && t._fileSize >= t._maxSize*int64(t._unit)
	}
	return false
}

func (t *fileObj) rename() (bckupfilename string, err error) {
	if t._rolltype == _DAYLY {
		bckupfilename = getBackupDayliFileName(t._fileDir, t._fileName, t._mode, t._gzip)
	} else {
		bckupfilename, err = getBackupRollFileName(t._fileDir, t._fileName, t._gzip)
	}
	if bckupfilename != "" && err == nil {
		oldPath := fmt.Sprint(t._fileDir, "/", t._fileName)
		newPath := fmt.Sprint(t._fileDir, "/", bckupfilename)
		err = os.Rename(oldPath, newPath)
		go func() {
			if err == nil && t._gzip {
				if err = lgzip(fmt.Sprint(newPath, ".gz"), bckupfilename, newPath); err == nil {
					os.Remove(newPath)
				}
			}
			if err == nil && t._rolltype == _ROLLFILE && t._maxFileNum > 0 {
				_rmOverCountFile(t._fileDir, bckupfilename, t._maxFileNum, t._gzip)
			}
		}()
	}
	return
}

func (t *fileObj) close() (err error) {
	defer catchError()
	if t._fileHandler != nil {
		err = t._fileHandler.Close()
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
		buf.Free()
	} else {
		fmt.Println(string(s))
	}
}

func k1(calldepth int) int {
	return calldepth + 1
}

func getOutBuffer(s []byte, levelname []byte, flag _FORMAT, calldepth int) (buf *Buffer) {
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

var m = NewLimitMap[any, runtime.Frame](1 << 13)

func output(flag _FORMAT, calldepth int, s []byte, levelname []byte) (buf *Buffer) {
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
	buf = NewBufferByPool()
	formatHeader(buf, now, file, line, flag, levelname)
	buf.Write(s)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		buf.WriteByte('\n')
	}
	return
}

func formatHeader(buf *Buffer, t time.Time, file *string, line *int, flag _FORMAT, levelname []byte) {
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
			short := *file
			for i := len(*file) - 1; i > 0; i-- {
				if (*file)[i] == '/' {
					short = (*file)[i+1:]
					break
				}
			}
			file = &short
		}
		buf.Write([]byte(*file))
		buf.WriteByte(':')
		itoa(buf, *line, -1)
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

func itoa(buf *Buffer, i int, wid int) {
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
