// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logger

import (
	"github.com/donnie4w/gofer/buffer"
	"runtime"
)

type callStack struct {
	stack []callerInfo
}

type callerInfo struct {
	FileName string
	Line     int
	FuncName string
}

func (cs *callStack) Push(fileName string, line int) {
	cs.stack = append(cs.stack, callerInfo{FileName: fileName, Line: line})
}

func (cs *callStack) PushWithFunc(fileName string, line int, funcName string) {
	cs.stack = append(cs.stack, callerInfo{FileName: fileName, Line: line, FuncName: funcName})
}

func (cs *callStack) Pop(flag _FORMAT, fileBuffer *buffer.Buffer) {
	for i, ci := range cs.stack {
		getfileInfo(&flag, &ci.FileName, &ci.Line, &ci.FuncName, fileBuffer)
		if i < len(cs.stack)-1 {
			fileBuffer.WriteByte('#')
		}
	}
	return
}

func collectCallStack(depth int, formatfunc bool, stack *callStack, recursion bool) *callStack {
	if depth <= 0 {
		return stack
	}
	if stack == nil {
		stack = &callStack{make([]callerInfo, 0)}
	}
	i := 1
	for {
		var pcs [1]uintptr
		if more := runtime.Callers(depth+i, pcs[:]); more == 0 {
			return stack
		}
		var f runtime.Frame
		var ok bool
		if f, ok = m.Get(pcs); !ok {
			f, _ = runtime.CallersFrames([]uintptr{pcs[0]}).Next()
			m.Put(pcs, f)
		}
		if formatfunc {
			stack.PushWithFunc(f.File, f.Line, funcname(f.Func.Name()))
		} else {
			stack.Push(f.File, f.Line)
		}
		if !recursion {
			break
		}
		i++
	}

	return stack
}
