// Copyright (c) 2014, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/go-logger

package logger

import (
	"github.com/donnie4w/gofer/buffer"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var _current = ""
var _lastUpdate time.Time

func getfileInfo(flag *_FORMAT, fileName *string, line *int, funcName *string, filebuf *buffer.Buffer) {
	if *flag&(FORMAT_SHORTFILENAME|FORMAT_LONGFILENAME|FORMAT_RELATIVEFILENAME) != 0 {
		if *flag&FORMAT_SHORTFILENAME != 0 {
			short := *fileName
			for i := len(*fileName) - 1; i > 0; i-- {
				if (*fileName)[i] == '/' {
					short = (*fileName)[i+1:]
					break
				}
			}
			fileName = &short
		} else if *flag&FORMAT_RELATIVEFILENAME != 0 {
			if time.Since(_lastUpdate) > time.Second || _current == "" {
				if c, err := os.Getwd(); err == nil {
					_current = c
				}
			}
			_lastUpdate = time.Now()
			if _current != "" {
				if relative, err := filepath.Rel(_current, *fileName); err == nil {
					fileName = &relative
				}
			}
		}
		filebuf.Write([]byte(*fileName))
		if *flag&FORMAT_FUNC != 0 && funcName != nil && *funcName != "" {
			filebuf.WriteByte(':')
			filebuf.WriteString(*funcName)
		}
		filebuf.WriteByte(':')
		filebuf.Write(itoa(*line, -1))
	}
}

func funcname(str string) string {
	if lastDotIndex := strings.LastIndex(str, "."); lastDotIndex != -1 {
		return str[lastDotIndex+1:]
	}
	return str
}
