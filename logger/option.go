// Copyright (c) 2014, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/go-logger

package logger

// FileOption defines the configuration interface for log file rotation.
// It provides settings for file rotation mode, time-based rotation, file path,
// maximum file size, maximum backup count, and compression options.
// FileOption 定义了日志文件切割的配置接口，
// 提供了文件切割模式、时间切割、文件路径、最大文件大小、最大备份数和压缩选项等配置项。
type FileOption interface {
	Cutmode() _CUTMODE    // Returns the file rotation mode type / 返回文件切割模式类型
	TimeMode() _MODE_TIME // Returns the time-based rotation mode / 返回时间切割模式
	FilePath() string     // Returns the log file path / 返回日志文件路径
	MaxSize() int64       // Returns the maximum file size (in bytes) / 返回最大文件大小（字节）
	MaxBackup() int       // Returns the maximum number of backup files / 返回最大备份文件数量
	Compress() bool       // Returns whether compression is enabled / 返回是否启用压缩
}

// FileSizeMode defines the configuration for file rotation based on file size.
// FileSizeMode 定义了按文件大小切割的配置。
type FileSizeMode struct {
	Filename   string // The path to the log file / 日志文件路径
	Maxsize    int64  // The maximum file size; when exceeded, a rotation will occur / 文件最大大小，超过该大小时进行切割
	Maxbackup  int    // The maximum number of backup files to keep / 保留的最大备份文件数量
	IsCompress bool   // Whether to enable compression for backup files / 是否启用备份文件的压缩
}

func (f *FileSizeMode) Cutmode() _CUTMODE {
	return _SIZEMODE // Indicates rotation by file size / 表示按文件大小进行切割
}

func (f *FileSizeMode) TimeMode() _MODE_TIME {
	return MODE_HOUR // This function is not used in this mode
}

func (f *FileSizeMode) FilePath() string {
	return f.Filename // Returns the log file path / 返回日志文件路径
}

func (f *FileSizeMode) MaxSize() int64 {
	return f.Maxsize // Returns the maximum file size limit / 返回最大文件大小限制
}

func (f *FileSizeMode) MaxBackup() int {
	return f.Maxbackup // Returns the maximum number of backup files / 返回最大备份文件数量
}

func (f *FileSizeMode) Compress() bool {
	return f.IsCompress // Returns whether compression is enabled / 返回是否启用压缩
}

// FileTimeMode defines the configuration for file rotation based on time.
// FileTimeMode 定义了按时间切割的配置。
type FileTimeMode struct {
	Filename   string     // The path to the log file / 日志文件路径
	Timemode   _MODE_TIME // The time-based rotation mode / 时间切割模式
	Maxbackup  int        // The maximum number of backup files to keep / 保留的最大备份文件数量
	IsCompress bool       // Whether to enable compression for backup files / 是否启用备份文件的压缩
}

func (f *FileTimeMode) Cutmode() _CUTMODE {
	return _TIMEMODE // Indicates rotation by time / 表示按时间切割
}

func (f *FileTimeMode) TimeMode() _MODE_TIME {
	return f.Timemode // Returns the time-based rotation mode / 返回时间切割模式
}

func (f *FileTimeMode) FilePath() string {
	return f.Filename // Returns the log file path / 返回日志文件路径
}

func (f *FileTimeMode) MaxSize() int64 {
	return 0 // No size limitation for time-based rotation / 按时间切割不考虑文件大小限制
}

func (f *FileTimeMode) MaxBackup() int {
	return f.Maxbackup // Returns the maximum number of backup files / 返回最大备份文件数量
}

func (f *FileTimeMode) Compress() bool {
	return f.IsCompress // Returns whether compression is enabled / 返回是否启用压缩
}

// FileMixedMode defines the configuration for file rotation based on both time and file size.
// FileMixedMode 定义了按时间和文件大小混合切割的配置。
type FileMixedMode struct {
	Filename   string     // The path to the log file / 日志文件路径
	Timemode   _MODE_TIME // The time-based rotation mode / 时间切割模式
	Maxsize    int64      // The maximum file size; when exceeded, a rotation will occur / 文件最大大小，超过该大小时进行切割
	Maxbackup  int        // The maximum number of backup files to keep / 保留的最大备份文件数量
	IsCompress bool       // Whether to enable compression for backup files / 是否启用备份文件的压缩
}

func (f *FileMixedMode) Cutmode() _CUTMODE {
	return _MIXEDMODE // Indicates rotation by both time and size (mixed mode) / 表示按时间和大小进行混合切割
}

func (f *FileMixedMode) TimeMode() _MODE_TIME {
	return f.Timemode // Returns the time-based rotation mode / 返回时间切割模式
}

func (f *FileMixedMode) FilePath() string {
	return f.Filename // Returns the log file path / 返回日志文件路径
}

func (f *FileMixedMode) MaxSize() int64 {
	return f.Maxsize // Returns the maximum file size for rotation / 返回文件最大大小
}

func (f *FileMixedMode) MaxBackup() int {
	return f.Maxbackup // Returns the maximum number of backup files / 返回最大备份文件数量
}

func (f *FileMixedMode) Compress() bool {
	return f.IsCompress // Returns whether compression is enabled / 返回是否启用压缩
}

// Option represents a configuration option for the Logging struct.
// It includes various settings such as log level, console output, format, formatter, file options, and a custom handler.
type Option struct {
	Level      LEVELTYPE  // Log level, e.g., DEBUG, INFO, WARN, ERROR, etc.
	Console    bool       // Whether to also output logs to the console.
	Format     _FORMAT    // Log format.
	Formatter  string     // Formatting string for customizing the log output format.
	FileOption FileOption // File-specific options for the log handler.
	Stacktrace LEVELTYPE  // Log level, e.g., DEBUG, INFO, WARN, ERROR, etc.
	// CustomHandler
	//
	// When customHandler returns false, the println function returns without executing further prints. Returning true allows the subsequent print operations to continue.
	//
	// customHandler返回false时，println函数返回，不再执行后续的打印，返回true时，继续执行后续打印。
	CustomHandler func(lc *LogContext) bool // Custom log handler function allowing users to define additional log processing logic.

	// AttrFormat defines a set of customizable formatting functions for log entries.
	// This structure allows users to specify custom formats for log levels, timestamps,
	// and message bodies, enabling highly flexible log formatting.
	AttrFormat *AttrFormat

	// CallDepth Custom function call depth
	CallDepth int
}

type LogContext struct {
	Level LEVELTYPE
	Args  []any
}

type LevelOption struct {
	Format    _FORMAT // Log format.
	Formatter string  // Formatting string for customizing the log output format.
}

// AttrFormat defines a set of customizable formatting functions for log entries.
// This structure allows users to specify custom formats for log levels, timestamps,
// and message bodies, enabling highly flexible log formatting.
//
// Example usage:
//
//	attrFormat := &AttrFormat{
//	    SetLevelFmt: func(level LEVELTYPE) string {
//	        switch level {
//	        case LEVEL_DEBUG:
//	            return "DEBUG:"
//	        case LEVEL_INFO:
//	            return "INFO :"
//	        case LEVEL_WARN:
//	            return "WARN :"
//	        case LEVEL_ERROR:
//	            return "ERROR:"
//	        case LEVEL_FATAL:
//	            return "FATAL:"
//	        default:
//	            return "UNKNOWN:"
//	        }
//	    },
//	    SetTimeFmt: func() (string, string, string) {
//	        now := time.Now().Format("2006-01-02 15:04:05")
//	        return now, "", ""
//	    },
//	    SetBodyFmt: func(level LEVELTYPE, msg []byte) []byte {
//	        switch level {
//	        case LEVEL_DEBUG:
//	            return append([]byte("\033[34m"), append(msg, '\033', '[', '0', 'm')...) // Blue for DEBUG
//	        case LEVEL_INFO:
//	            return append([]byte("\033[32m"), append(msg, '\033', '[', '0', 'm')...) // Green for INFO
//	        case LEVEL_WARN:
//	            return append([]byte("\033[33m"), append(msg, '\033', '[', '0', 'm')...) // Yellow for WARN
//	        case LEVEL_ERROR:
//	            return append([]byte("\033[31m"), append(msg, '\033', '[', '0', 'm')...) // Red for ERROR
//	        case LEVEL_FATAL:
//	            return append([]byte("\033[41m"), append(msg, '\033', '[', '0', 'm')...) // Red background for FATAL
//	        default:
//	            return msg
//	        }
//	    },
//	}
type AttrFormat struct {
	// SetLevelFmt defines a function to format log levels.
	// This function receives a log level of type LEVELTYPE and returns a formatted string.
	// The string represents the level prefix in log entries, such as "DEBUG:" for debug level or "FATAL:" for fatal level.
	//
	// Example:
	//   SetLevelFmt: func(level LEVELTYPE) string {
	//       if level == LEVEL_DEBUG {
	//           return "DEBUG:"
	//       }
	//       return "INFO:"
	//   }
	SetLevelFmt func(level LEVELTYPE) string

	// SetTimeFmt defines a function to format timestamps for log entries.
	// This function returns three strings representing different components of a timestamp.
	// It allows custom handling of dates, times, or other time-based information.
	//
	// Example:
	//   SetTimeFmt: func() (string, string, string) {
	//       currentTime := time.Now().Format("2006-01-02 15:04:05")
	//       return currentTime, "", ""
	//   }
	SetTimeFmt func() (string, string, string)

	// SetBodyFmt defines a function to customize the format of log message bodies.
	// This function receives the log level and the message body in byte slice format, allowing
	// modifications such as adding colors, handling line breaks, or appending custom suffixes.
	//
	// Example:
	//   SetBodyFmt: func(level LEVELTYPE, msg []byte) []byte {
	//       if level == LEVEL_ERROR {
	//           return append([]byte("ERROR MESSAGE: "), msg...)
	//       }
	//       return msg
	//   }
	SetBodyFmt func(level LEVELTYPE, msg []byte) []byte
}
