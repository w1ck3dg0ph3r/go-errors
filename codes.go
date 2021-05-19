package errors

// ErrorKind is an error's kind
//
// Error can have multiple kinds specified at once.
type ErrorKind int

// ErrorCode is an error's type code
//
// This package defines some common error codes, but generally you should
// user error codes relevant to your application.
type ErrorCode int

const (
	Client ErrorKind = 1 << iota
	Server
	Transient
)

const (
	Unexpected ErrorCode = iota
	Invalid
	IO
	Deadlock
	Permission
	AlreadyExists
	NotFound
)
