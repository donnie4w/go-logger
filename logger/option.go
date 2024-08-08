// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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

// Option represents a configuration option for the Logging struct.
// It includes various settings such as log level, console output, format, formatter, file options, and a custom handler.
type Option struct {
	Level      _LEVEL     // Log level, e.g., DEBUG, INFO, WARN, ERROR, etc.
	Console    bool       // Whether to also output logs to the console.
	Format     _FORMAT    // Log format.
	Formatter  string     // Formatting string for customizing the log output format.
	FileOption FileOption // File-specific options for the log handler.
	Stacktrace _LEVEL     // Log level, e.g., DEBUG, INFO, WARN, ERROR, etc.
	// CustomHandler
	//
	// When customHandler returns false, the println function returns without executing further prints. Returning true allows the subsequent print operations to continue.
	//
	// customHandler返回false时，println函数返回，不再执行后续的打印，返回true时，继续执行后续打印。
	CustomHandler func(lc *LogContext) bool // Custom log handler function allowing users to define additional log processing logic.
}

type LogContext struct {
	Level _LEVEL
	Args  []any
}

type LevelOption struct {
	Format    _FORMAT // Log format.
	Formatter string  // Formatting string for customizing the log output format.
}
