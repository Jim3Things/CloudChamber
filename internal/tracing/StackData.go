// This module contains a set of helper functions for pulling data from the
// current stack.  Specifically, it includes getting a formatted stack trace
// and getting a specific stack frame's method name.

package tracing

import (
	"fmt"
	"runtime"
	"strings"
)

const (
	stackDepth = 5
)

// MethodName returns the caller's method name, without the leading directory paths.
func MethodName(skip int) string {
	addresses := make([]uintptr, 1)

	// Get the information up the stack (i.e. the caller of this method, or beyond)
	runtime.Callers(skip + 1, addresses)
	frames := runtime.CallersFrames(addresses)
	frame, _ := frames.Next()

	// ... and return the name
	name := frame.Func.Name()

	idx := strings.LastIndex(name, "/")
	if idx >= 0 {
		name = name[idx + 1:]
	}

	return name
}

// StackTrace produces a formatted call stack, in the form of filename and line
// number.  A newline splits each entry.
func StackTrace() string {
	res := ""

	addresses := make([]uintptr, stackDepth)
	runtime.Callers(1, addresses)
	frames := runtime.CallersFrames(addresses)

	frame, more := frames.Next()
	res = fmt.Sprintf("%s:%d", frame.File, frame.Line)

	for more {
		frame, more = frames.Next()
		res = fmt.Sprintf("%s\n%s:%d", res, frame.File, frame.Line)
	}

	return res
}
