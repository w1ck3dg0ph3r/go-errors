// Package errors provides error wrapping with stack trace
package errors

import (
	stderr "errors"
	"strings"
)

// Op is an operation during which an error has occurred.
type Op string

// Error is an error wrapper.
type Error struct {
	Op    Op
	Kind  ErrorKind
	Code  ErrorCode
	Msg   string
	Cause error
	Stack StackTrace
}

// Error return human readable representation of an error.
func (e *Error) Error() string {
	sb := &strings.Builder{}
	if e.Msg != "" {
		sb.WriteString(e.Msg)
	}
	if e.Cause != nil {
		causeMsg := e.Cause.Error()
		if e.Msg != "" && causeMsg != "" {
			sb.WriteString(": ")
		}
		sb.WriteString(causeMsg)
	}
	return sb.String()
}

// E creates or wraps an error.
// Arguments could be an Op, ErrorKind, ErrorCode, string message, or an error to wrap.
func E(args ...interface{}) *Error {
	e := &Error{}
	wrapping := false
	for _, a := range args {
		switch a := a.(type) {
		case Op:
			e.Op = a
		case ErrorKind:
			e.Kind |= a
		case ErrorCode:
			e.Code = a
		case string:
			e.Msg = a
		case error:
			e.Cause = a
			wrapping = true
		case nil:
			return nil
		default:
			panic("bad call to E")
		}
	}
	shouldTrace := true
	if wrapping {
		if _, ok := e.Cause.(*Error); ok {
			shouldTrace = false
		}
	}
	if shouldTrace {
		e.Stack = callers()
	}
	return e
}

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
func Unwrap(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return nil
	}
	return u.Unwrap()
}

// Unwrap unwraps an error
func (e *Error) Unwrap() error {
	return e.Cause
}

// Ops returns stack of error operations.
func Ops(err error) []Op {
	e, ok := err.(*Error)
	if !ok {
		return []Op{}
	}
	res := []Op{e.Op}
	cause, ok := e.Cause.(*Error)
	if !ok {
		return res
	}
	res = append(res, Ops(cause)...)
	return res
}

// Trace returns error's stack trace.
func Trace(err error) StackTrace {
	e, ok := err.(*Error)
	if !ok {
		return nil
	}
	cause, ok := e.Cause.(*Error)
	if !ok {
		return e.Stack
	}
	return Trace(cause)
}

// Kind returns error's kind.
func Kind(err error) ErrorKind {
	e, ok := err.(*Error)
	if !ok {
		return 0
	}
	if e.Kind != 0 {
		return e.Kind
	}
	if e.Cause != nil {
		return Kind(e.Cause)
	}
	return 0
}

// Code returns error's code.
func Code(err error) ErrorCode {
	e, ok := err.(*Error)
	if !ok {
		return Unexpected
	}
	if e.Code != 0 {
		return e.Code
	}
	if e.Cause != nil {
		return Code(e.Cause)
	}
	return Unexpected
}

// Is checks if err is of given kind or has given code
func Is(err error, what interface{}) bool {
	e, ok := err.(*Error)
	if !ok {
		if code, ok := what.(ErrorCode); ok {
			return code == Unexpected
		}
		return false
	}
	switch what := what.(type) {
	case ErrorKind:
		if e.Kind != 0 {
			return e.Kind&what > 0
		}
	case ErrorCode:
		if e.Code != 0 {
			return e.Code == what
		}
	}
	if e.Cause != nil {
		return Is(e.Cause, what)
	}
	return false
}

// IsAnyOf checks if err is any of the given kinds or has any og the given codes
func IsAnyOf(err error, what ...interface{}) bool {
	for i := range what {
		if Is(err, what[i]) {
			return true
		}
	}
	return false
}

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true. Otherwise, it returns false.
func As(err error, target interface{}) bool {
	return stderr.As(err, target)
}

// ClientMsg returns error message suitable to display to the client.
func ClientMsg(err error) string {
	e, ok := err.(*Error)
	if !ok {
		return ""
	}
	if e.Kind&Client > 0 {
		return e.Msg
	}
	if e.Cause != nil {
		return ClientMsg(e.Cause)
	}
	return ""
}
