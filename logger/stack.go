// Copyright (c) 2014, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/go-logger

package logger

import (
	"github.com/donnie4w/gofer/buffer"
	pool "github.com/donnie4w/gofer/pool/buffer"
	"runtime"
)

type callStack struct {
	stack []*callerInfo
}

func (c *callStack) reset() {
	if len(c.stack) > 0 {
		c.stack = c.stack[:0]
	}
}

var callStackPool = pool.NewPool[callStack](func() *callStack {
	return &callStack{stack: make([]*callerInfo, 0)}
}, func(c *callStack) {
	c.reset()
})

type callerInfo struct {
	FileName string
	Line     int
	FuncName string
}

func (c *callerInfo) reset() {
	c.FileName, c.FuncName = "", ""
}

var callerInfoPool = pool.NewPool[callerInfo](func() *callerInfo {
	return &callerInfo{}
}, func(c *callerInfo) {
	c.reset()
})

func (cs *callStack) Push(fileName string, line int) {
	ci := callerInfoPool.Get()
	ci.FileName, ci.Line = fileName, line
	cs.stack = append(cs.stack, ci)
}

func (cs *callStack) PushWithFunc(fileName string, line int, funcName string) {
	ci := callerInfoPool.Get()
	ci.FileName, ci.Line, ci.FuncName = fileName, line, funcName
	cs.stack = append(cs.stack, ci)
}

func (cs *callStack) Pop(flag _FORMAT, fileBuffer *buffer.Buffer) {
	for i, ci := range cs.stack {
		getfileInfo(&flag, &ci.FileName, &ci.Line, &ci.FuncName, fileBuffer)
		if i < len(cs.stack)-1 {
			fileBuffer.WriteByte('#')
		}
		callerInfoPool.Put(&ci)
	}
	return
}

func collectCallStack(depth int, formatfunc bool, stack *callStack, recursion bool) *callStack {
	if depth <= 0 {
		return stack
	}
	if stack == nil {
		stack = callStackPool.Get()
	}
	i := 1
	for {
		var pcs [1]uintptr
		if more := runtime.Callers(depth+i, pcs[:]); more == 0 {
			return stack
		}
		f, ok := m.Get(pcs[0])
		if !ok {
			f, _ = runtime.CallersFrames([]uintptr{pcs[0]}).Next()
			m.Put(pcs[0], f)
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

//func collectCallStack(depth int, formatfunc bool, stack *callStack, recursion bool) *callStack {
//	if depth <= 0 {
//		return stack
//	}
//
//	pcs := make([]uintptr, 0, 8)
//	n := runtime.Callers(depth+1, pcs[:cap(pcs)])
//	for n == cap(pcs) {
//		newCap := cap(pcs) + 8
//		pcs = make([]uintptr, 0, newCap)
//		n = runtime.Callers(depth+1, pcs[:cap(pcs)])
//	}
//
//	if stack == nil {
//		stack = callStackPool.Get()
//	}
//
//	pcs = pcs[:n]
//	for i := 0; i < n; i++ {
//		f, ok := m.Get(pcs[i])
//		if !ok {
//			f, _ = runtime.CallersFrames(pcs[i : i+1]).Next()
//			m.Put(pcs[i], f)
//		}
//		if formatfunc {
//			stack.PushWithFunc(f.File, f.Line, funcname(f.Func.Name()))
//		} else {
//			stack.Push(f.File, f.Line)
//		}
//		if !recursion {
//			break
//		}
//	}
//	return stack
//}
