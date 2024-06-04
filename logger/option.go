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

type Option struct {
	Level      _LEVEL
	Console    bool
	Format     _FORMAT
	Formatter  string
	FileOption FileOption
}
