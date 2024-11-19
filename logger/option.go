// Copyright (c) 2014, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/go-logger

package logger

type FileOption interface {
	Cutmode() _CUTMODE
	TimeMode() _MODE_TIME
	FilePath() string
	MaxSize() uint64
	MaxBuckup() int
	Compress() bool
}

type FileSizeMode struct {
	Filename   string
	Maxsize    uint64
	Maxbuckup  int
	IsCompress bool
}

func (f *FileSizeMode) Cutmode() _CUTMODE {
	return _SIZEMODE
}

func (f *FileSizeMode) TimeMode() _MODE_TIME {
	return MODE_HOUR
}

func (f *FileSizeMode) FilePath() string {
	return f.Filename
}
func (f *FileSizeMode) MaxSize() uint64 {
	return f.Maxsize
}
func (f *FileSizeMode) MaxBuckup() int {
	return f.Maxbuckup
}

func (f *FileSizeMode) Compress() bool {
	return f.IsCompress
}

type FileTimeMode struct {
	Filename   string
	Timemode   _MODE_TIME
	Maxbuckup  int
	IsCompress bool
}

func (f *FileTimeMode) Cutmode() _CUTMODE {
	return _TIMEMODE
}

func (f *FileTimeMode) TimeMode() _MODE_TIME {
	return f.Timemode
}

func (f *FileTimeMode) FilePath() string {
	return f.Filename
}
func (f *FileTimeMode) MaxSize() uint64 {
	return 0
}
func (f *FileTimeMode) MaxBuckup() int {
	return f.Maxbuckup
}

func (f *FileTimeMode) Compress() bool {
	return f.IsCompress
}

type FileMixedMode struct {
	Filename   string
	Timemode   _MODE_TIME
	Maxsize    uint64
	SizeUint   _UNIT
	Maxbuckup  int
	IsCompress bool
}

func (f *FileMixedMode) Cutmode() _CUTMODE {
	return _MIXEDMODE
}

func (f *FileMixedMode) TimeMode() _MODE_TIME {
	return f.Timemode
}

func (f *FileMixedMode) FilePath() string {
	return f.Filename
}
func (f *FileMixedMode) MaxSize() uint64 {
	return f.Maxsize * uint64(f.SizeUint)
}
func (f *FileMixedMode) MaxBuckup() int {
	return f.Maxbuckup
}

func (f *FileMixedMode) Compress() bool {
	return f.IsCompress
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

	AttrFormat *AttrFormat
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
