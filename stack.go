package errors

// Heavily based on https://github.com/xpsuper/stl/blob/master/stl.stack.go

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strconv"
	"strings"
)

// StackFrame represents a program counter inside a Trace frame.
// For historical reasons if StackFrame is interpreted as a uintptr
// its value represents the program counter + 1.
type StackFrame uintptr

// pc returns the program counter for this frame;
// multiple frames may have the same PC value.
func (f StackFrame) pc() uintptr { return uintptr(f) - 1 }

// file returns the full path to the file that contains the
// function for this StackFrame's pc.
func (f StackFrame) file() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return unknownFunction
	}
	file, _ := fn.FileLine(f.pc())
	return file
}

// line returns the line number of source code of the
// function for this StackFrame's pc.
func (f StackFrame) line() int {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return 0
	}
	_, line := fn.FileLine(f.pc())
	return line
}

// name returns the name of this function, if known.
func (f StackFrame) name() string {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return unknownFunction
	}
	return fn.Name()
}

// Format formats the frame according to the fmt.Formatter interface.
//
//    %s    source file
//    %d    source line
//    %n    function name
//    %v    equivalent to %s:%d
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//    %+s   function name and path of source file relative to the compile time
//          GOPATH separated by \n\t (<funcname>\n\t<path>)
//    %+v   equivalent to %+s:%d
//nolint:errcheck
func (f StackFrame) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			io.WriteString(s, f.name())
			io.WriteString(s, "\n\t")
			io.WriteString(s, f.file())
		default:
			io.WriteString(s, path.Base(f.file()))
		}
	case 'd':
		io.WriteString(s, strconv.Itoa(f.line()))
	case 'n':
		io.WriteString(s, funcname(f.name()))
	case 'v':
		f.Format(s, 's')
		io.WriteString(s, ":")
		f.Format(s, 'd')
	}
}

// MarshalText formats a StackTrace StackFrame as a text string. The output is the
// same as that of fmt.Sprintf("%+v", f), but without newlines or tabs.
func (f StackFrame) MarshalText() ([]byte, error) {
	name := f.name()
	if name == unknownFunction {
		return []byte(name), nil
	}
	return []byte(fmt.Sprintf("%s %s:%d", name, f.file(), f.line())), nil
}

// StackTrace is Trace of Frames from innermost (newest) to outermost (oldest).
type StackTrace []StackFrame

// Format formats the Trace of Frames according to the fmt.Formatter interface.
//
//    %s	lists source files for each StackFrame in the Trace
//    %v	lists the source file and line number for each StackFrame in the Trace
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//    %+v   Prints filename, function, and line number for each StackFrame in the Trace.
//nolint:errcheck
func (st StackTrace) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('+'):
			for _, f := range st {
				io.WriteString(s, "\n")
				f.Format(s, verb)
			}
		case s.Flag('#'):
			fmt.Fprintf(s, "%#v", []StackFrame(st))
		default:
			st.formatSlice(s, verb)
		}
	case 's', 'n':
		st.formatSlice(s, verb)
	}
}

// formatSlice will format this StackTrace into the given buffer as a slice of
// StackFrame, only valid when called with '%s' or '%v'.
//nolint:errcheck
func (st StackTrace) formatSlice(s fmt.State, verb rune) {
	io.WriteString(s, "[")
	for i, f := range st {
		if i > 0 {
			io.WriteString(s, " ")
		}
		f.Format(s, verb)
	}
	io.WriteString(s, "]")
}

func callers() StackTrace {
	const depth = 32
	const framesToSkip = 3
	var pcs [depth]uintptr
	n := runtime.Callers(framesToSkip, pcs[:])
	st := make(StackTrace, n)
	for i := 0; i < n; i++ {
		st[i] = StackFrame(pcs[i])
	}
	return st
}

// funcname removes the path prefix component of a function's name reported by func.Name().
func funcname(name string) string {
	i := strings.LastIndex(name, "/")
	name = name[i+1:]
	i = strings.Index(name, ".")
	return name[i+1:]
}

const unknownFunction = "unknown"
