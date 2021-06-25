package errors

import (
	stderr "errors"
)

// List is an error type that can hold multiple errors.
// It could be used to return accumulated errors from the function
// as a single error.
type List []error

// Multiple returns slice of errors if err is a List, slice with one error err otherwise.
func Multiple(err error) []error {
	if list, ok := err.(List); ok {
		return list
	}
	return []error{err}
}

// Has checks if err contains an error of given ErrorKind, with given ErrorCode or matches given error target.
func Has(err error, target interface{}) bool {
	list, ok := err.(List)
	if !ok {
		return Is(err, target)
	}
	if len(list) == 0 {
		return false
	}
	for i := range list {
		if Is(list[i], target) {
			return true
		}
	}
	return false
}

// HasAnyOf checks if any error in err
func HasAnyOf(err error, what ...interface{}) bool {
	if err == nil {
		return false
	}
	if list, ok := err.(List); ok && len(list) == 0 {
		return false
	}
	for i := range what {
		if Has(err, what[i]) {
			return true
		}
	}
	return false
}

// ErrOrNil return nil if error list is empty, otherwise list itself.
// This is useful to eliminate check for an empty list when returning from function.
func (l List) ErrOrNil() error {
	if len(l) == 0 {
		return nil
	}
	return l
}

// Add adds an error to list.
func (l *List) Add(e error) {
	if e != nil {
		*l = append(*l, e)
	}
}

// Clear removes all errors from list.
func (l *List) Clear() {
	*l = (*l)[:0]
}

func (l List) Unwrap() error {
	if l == nil || len(l) == 0 {
		return nil
	}
	if len(l) == 1 {
		return l[0]
	}
	rest := make(List, len(l)-1)
	copy(rest, l[1:])
	return &rest
}

func (l List) Is(target error) bool {
	if l == nil || len(l) == 0 {
		return false
	}
	return stderr.Is(l[0], target)
}

func (l List) As(target interface{}) bool {
	if l == nil || len(l) == 0 {
		return false
	}
	return stderr.As(l[0], target)
}

// Errors returns human readable representation of first error in the list
// or empty string if list is empty.
func (l List) Error() string {
	if l == nil || len(l) == 0 {
		return ""
	}
	return l[0].Error()
}
