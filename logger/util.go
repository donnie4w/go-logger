// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logger

import (
	"github.com/donnie4w/gofer/buffer"
	"strings"
)

func getfileInfo(flag *_FORMAT, fileName *string, line *int, funcName *string, filebuf *buffer.Buffer) {
	if *flag&(FORMAT_SHORTFILENAME|FORMAT_LONGFILENAME) != 0 {
		if *flag&FORMAT_SHORTFILENAME != 0 {
			short := *fileName
			for i := len(*fileName) - 1; i > 0; i-- {
				if (*fileName)[i] == '/' {
					short = (*fileName)[i+1:]
					break
				}
			}
			fileName = &short
		}
		filebuf.Write([]byte(*fileName))
		if *flag&FORMAT_FUNC != 0 && funcName != nil && *funcName != "" {
			filebuf.WriteByte(':')
			filebuf.WriteString(*funcName)
		}
		filebuf.WriteByte(':')
		itoa(filebuf, *line, -1)
	}
}

func funcname(str string) string {
	if lastDotIndex := strings.LastIndex(str, "."); lastDotIndex != -1 {
		return str[lastDotIndex+1:]
	}
	return str
}
