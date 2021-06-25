package errors

// ErrorKind is an error's kind
//
// Error can have multiple kinds specified at once.
type ErrorKind int

// ErrorCode is an error's type code
//
// This package defines some common error codes, but generally you should
// use error codes relevant to your application.
type ErrorCode int

// Predefined error kinds
const (
	Client ErrorKind = 1 << iota
	Server
	Transient
)

// Predefined error codes
const (
	Unexpected ErrorCode = iota
	Invalid
	IO
	Deadlock
	Permission
	AlreadyExists
	NotFound
)
