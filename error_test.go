package errors_test

import (
	"encoding/json"
	stderr "errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/w1ck3dg0ph3r/go-errors"
)

func Test_E(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		err := errors.E(nil)
		assert.Nil(t, err)
	})

	t.Run("all arguments", func(t *testing.T) {
		cause := fmt.Errorf("cause")
		err := errors.E(errors.Op("op"), errors.Server, errors.Transient, errors.IO, "msg", cause)
		assert.Equal(t, errors.Op("op"), err.Op)
		assert.Equal(t, errors.Server|errors.Transient, err.Kind)
		assert.Equal(t, errors.IO, err.Code)
		assert.Equal(t, "msg", err.Msg)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("invalid arguments", func(t *testing.T) {
		type notAnError struct{}
		assert.PanicsWithValue(t, "bad call to E: argument of type int", func() {
			_ = errors.E(42)
		})
		assert.PanicsWithValue(t, "bad call to E: argument of type errors_test.notAnError", func() {
			_ = errors.E(notAnError{})
		})
	})

	t.Run("multiple ops", func(t *testing.T) {
		assert.PanicsWithValue(t, "bad call to E: multiple ops", func() {
			_ = errors.E(errors.Op("op1"), errors.Op("op2"))
		})
	})

	t.Run("multiple messages", func(t *testing.T) {
		assert.PanicsWithValue(t, "bad call to E: multiple messages", func() {
			_ = errors.E("msg1", "msg2")
		})
	})

	t.Run("multiple codes", func(t *testing.T) {
		assert.PanicsWithValue(t, "bad call to E: multiple codes", func() {
			_ = errors.E(errors.NotFound, errors.AlreadyExists)
		})
	})

	t.Run("multiple causes", func(t *testing.T) {
		assert.PanicsWithValue(t, "bad call to E: multiple causes", func() {
			err := fmt.Errorf("")
			_ = errors.E(err, "msg", errors.Invalid, err)
		})
	})
}

func Test_Kind(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		assert.Equal(t, errors.ErrorKind(0), errors.Kind(nil))
	})

	t.Run("string error", func(t *testing.T) {
		err := fmt.Errorf("msg")
		assert.Equal(t, errors.ErrorKind(0), errors.Kind(err))
	})

	t.Run("no kind", func(t *testing.T) {
		err := errors.E(errors.NotFound)
		assert.Equal(t, errors.ErrorKind(0), errors.Kind(err))
	})

	t.Run("single kind", func(t *testing.T) {
		err := errors.E(errors.Client)
		assert.Equal(t, errors.Kind(err), errors.Client)
	})

	t.Run("multiple kinds", func(t *testing.T) {
		kind := errors.Kind(errors.E(errors.Server, errors.Transient))
		assert.True(t, kind&errors.Server > 0)
		assert.True(t, kind&errors.Transient > 0)
		assert.False(t, kind&errors.Client > 0)
	})

	t.Run("wrapped error with kind", func(t *testing.T) {
		err1 := errors.E("kind", errors.Transient)
		err2 := errors.E("kindless", err1)
		kind := errors.Kind(err2)
		assert.Equal(t, errors.Transient, kind)
	})

	t.Run("wrapped error, both with kind", func(t *testing.T) {
		err1 := errors.E("1", errors.Transient)
		err2 := errors.E(err1, "2", errors.Client)
		kind := errors.Kind(err2)
		assert.Equal(t, errors.Client, kind)
	})
}

func Test_Code(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		assert.Equal(t, errors.ErrorCode(0), errors.Code(nil))
	})

	t.Run("string error", func(t *testing.T) {
		assert.Equal(t, errors.Unexpected, errors.Code(fmt.Errorf("")))
	})

	t.Run("no code", func(t *testing.T) {
		assert.Equal(t, errors.Unexpected, errors.Code(errors.E("codeless")))
	})

	t.Run("error with code", func(t *testing.T) {
		err := errors.E(errors.IO)
		assert.Equal(t, errors.IO, errors.Code(err))
	})

	t.Run("wrapped error with code", func(t *testing.T) {
		err1 := errors.E(errors.NotFound)
		err2 := errors.E("wrapped", err1)
		assert.Equal(t, errors.NotFound, errors.Code(err2))
	})

	t.Run("wrapped error, both with code", func(t *testing.T) {
		err1 := errors.E(errors.NotFound)
		err2 := errors.E("wrapped", errors.IO, err1)
		assert.Equal(t, errors.IO, errors.Code(err2))
	})
}

func Test_Is(t *testing.T) {
	var notAnError struct{}
	var targetSomeError someError
	var targetError *errors.Error

	t.Run("nil", func(t *testing.T) {
		assert.False(t, errors.Is(nil, fs.ErrInvalid))
		assert.False(t, errors.Is(nil, errors.Server))
		assert.False(t, errors.Is(nil, errors.Unexpected))
	})

	t.Run("invalid target", func(t *testing.T) {
		err1 := errors.E("msg")
		err2 := fmt.Errorf("")
		assert.PanicsWithValue(t, "what must be ErrorKind, ErrorCode or error", func() {
			errors.Is(err1, 42)
		})
		assert.PanicsWithValue(t, "what must be ErrorKind, ErrorCode or error", func() {
			errors.Is(err2, 42)
		})
		assert.PanicsWithValue(t, "what must be ErrorKind, ErrorCode or error", func() {
			errors.Is(err1, notAnError)
		})
		assert.PanicsWithValue(t, "what must be ErrorKind, ErrorCode or error", func() {
			errors.Is(err2, notAnError)
		})
	})

	t.Run("error var", func(t *testing.T) {
		err := fs.ErrInvalid
		assert.True(t, errors.Is(err, fs.ErrInvalid))
		assert.True(t, errors.Is(err, errors.Unexpected))
		assert.False(t, errors.Is(err, fs.ErrNotExist))
		assert.False(t, errors.Is(err, errors.Server))
		assert.False(t, errors.Is(err, errors.Invalid))

		assert.True(t, stderr.Is(err, fs.ErrInvalid))
		assert.False(t, stderr.Is(err, fs.ErrNotExist))
	})

	t.Run("string error", func(t *testing.T) {
		err := fmt.Errorf("")
		assert.True(t, errors.Is(err, err))
		assert.True(t, errors.Is(err, errors.Unexpected))
		assert.False(t, errors.Is(err, errors.Server))
		assert.False(t, errors.Is(err, errors.IO))
		assert.False(t, errors.Is(err, targetSomeError))
		assert.False(t, errors.Is(err, targetError))

		assert.True(t, stderr.Is(err, err))
		assert.False(t, stderr.Is(err, targetSomeError))
		assert.False(t, stderr.Is(err, targetError))
	})

	t.Run("error with message only", func(t *testing.T) {
		var err error = errors.E("msg")
		assert.True(t, errors.Is(err, err))
		assert.False(t, errors.Is(err, errors.Client))
		assert.False(t, errors.Is(err, errors.Server))
		assert.False(t, errors.Is(err, errors.NotFound))
		assert.False(t, errors.Is(err, targetSomeError))

		assert.True(t, stderr.Is(err, err))
		assert.False(t, stderr.Is(err, targetSomeError))
	})

	t.Run("error with kind and code", func(t *testing.T) {
		var err error = errors.E(errors.Server, errors.IO)
		assert.True(t, errors.Is(err, err))
		assert.True(t, errors.Is(err, errors.Server))
		assert.True(t, errors.Is(err, errors.IO))
		assert.False(t, errors.Is(err, errors.Client))
		assert.False(t, errors.Is(err, errors.NotFound))

		assert.True(t, stderr.Is(err, err))
	})

	t.Run("wrapping error with kind and code", func(t *testing.T) {
		var err error = errors.E(errors.Op("op"), "msg", errors.Client, fs.ErrInvalid)
		assert.True(t, errors.Is(err, err))
		assert.True(t, errors.Is(err, fs.ErrInvalid))
		assert.True(t, errors.Is(err, errors.Unexpected))
		assert.True(t, errors.Is(err, errors.Client))
		assert.False(t, errors.Is(err, errors.Server))
		assert.False(t, errors.Is(err, errors.IO))

		assert.True(t, stderr.Is(err, err))
		assert.True(t, stderr.Is(err, fs.ErrInvalid))
	})

	t.Run("wrapped error, both with kind and code", func(t *testing.T) {
		var err1 error = errors.E(errors.Server, errors.IO)
		var err2 error = errors.E(errors.Client, errors.NotFound, err1)
		assert.True(t, errors.Is(err2, errors.Client))
		assert.True(t, errors.Is(err2, errors.NotFound))
		assert.True(t, errors.Is(err2, errors.Server))
		assert.True(t, errors.Is(err2, errors.IO))
		assert.True(t, errors.Is(err2, err1))
		assert.True(t, errors.Is(err2, err2))
		assert.False(t, errors.Is(err2, fs.ErrInvalid))
		assert.False(t, errors.Is(err2, errors.Transient))
		assert.False(t, errors.Is(err2, errors.Unexpected))

		assert.True(t, stderr.Is(err2, err1))
		assert.True(t, stderr.Is(err2, err2))
		assert.False(t, stderr.Is(err2, fs.ErrInvalid))
	})
}

func Test_IsAnyOf(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.False(t, errors.IsAnyOf(nil, errors.Server, errors.Transient, errors.IO))
	})

	t.Run("matches", func(t *testing.T) {
		err := errors.E(errors.Server, errors.IO)
		assert.True(t, errors.IsAnyOf(err, errors.Server, errors.Transient))
		assert.True(t, errors.IsAnyOf(err, errors.IO, errors.Deadlock))
	})

	t.Run("doesnt match", func(t *testing.T) {
		err := errors.E(errors.Server, errors.IO)
		assert.False(t, errors.IsAnyOf(err, errors.Client, errors.NotFound))
		assert.False(t, errors.IsAnyOf(err, errors.NotFound, errors.Invalid))
	})
}

func Test_As(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var res error
		assert.False(t, errors.As(nil, res))
	})

	t.Run("string error", func(t *testing.T) {
		err := fmt.Errorf("err1")
		var res1 error
		var res2 *errors.Error
		assert.True(t, errors.As(err, &res1))
		assert.Equal(t, "err1", res1.Error())
		assert.False(t, errors.As(err, &res2))
		assert.Nil(t, res2)

		assert.True(t, stderr.As(err, &res1))
		assert.Equal(t, "err1", res1.Error())
		assert.False(t, stderr.As(err, &res2))
		assert.Nil(t, res2)
	})

	t.Run("wrapped error", func(t *testing.T) {
		var err1 error = someError{code: 1234}
		var err2 error = errors.E(errors.Op("op1"), err1, errors.Client)
		var res1 *errors.Error
		var res2 someError

		assert.True(t, errors.As(err2, &res1))
		assert.Equal(t, errors.Client, res1.Kind)
		assert.Equal(t, errors.Op("op1"), res1.Op)
		assert.True(t, errors.As(err2, &res2))
		assert.Equal(t, 1234, res2.code)

		assert.True(t, stderr.As(err2, &res1))
		assert.Equal(t, errors.Client, res1.Kind)
		assert.Equal(t, errors.Op("op1"), res1.Op)
		assert.True(t, stderr.As(err2, &res2))
		assert.Equal(t, 1234, res2.code)
	})
}

func Test_Error_Message(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Equal(t, "", errors.ClientMsg(nil))
	})

	t.Run("string error", func(t *testing.T) {
		err := fmt.Errorf("msg1")
		assert.Equal(t, "msg1", err.Error())
		assert.Equal(t, "", errors.ClientMsg(err))
	})

	t.Run("error with msg", func(t *testing.T) {
		err := errors.E("msg1")
		assert.Equal(t, "msg1", err.Error())
		assert.Equal(t, "", errors.ClientMsg(err))
	})

	t.Run("error with client msg", func(t *testing.T) {
		err := errors.E("msg1", errors.Client)
		assert.Equal(t, "msg1", err.Error())
		assert.Equal(t, "msg1", errors.ClientMsg(err))
	})

	t.Run("wrapped error with msg", func(t *testing.T) {
		err1 := fmt.Errorf("msg1")
		err2 := errors.E("msg2", err1)
		assert.Equal(t, "msg2: msg1", err2.Error())
		assert.Equal(t, "", errors.ClientMsg(err2))
	})

	t.Run("wrapped error with client msg", func(t *testing.T) {
		err1 := errors.E("msg1", errors.Client)
		err2 := errors.E("msg2", err1)
		assert.Equal(t, "msg2: msg1", err2.Error())
		assert.Equal(t, "msg1", errors.ClientMsg(err2))
	})
}

func Test_Ops(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Empty(t, errors.Ops(nil))
	})

	t.Run("string error", func(t *testing.T) {
		assert.Empty(t, errors.Ops(fmt.Errorf("msg1")))
	})

	t.Run("error without op", func(t *testing.T) {
		assert.Empty(t, errors.Ops(errors.E("msg1")))
	})

	t.Run("error with op", func(t *testing.T) {
		err := errors.E(errors.Op("op1"), "msg1")
		assert.Equal(t, []errors.Op{"op1"}, errors.Ops(err))
	})

	t.Run("wrapped error with op", func(t *testing.T) {
		err1 := fmt.Errorf("msg1")
		err2 := errors.E(errors.Op("op2"), "msg2", err1)
		assert.Equal(t, []errors.Op{"op2"}, errors.Ops(err2))
	})

	t.Run("wrapped error, both with op", func(t *testing.T) {
		err1 := errors.E(errors.Op("op1"), "msg1")
		err2 := errors.E(errors.Op("op2"), "msg2", err1)
		assert.Equal(t, []errors.Op{"op2", "op1"}, errors.Ops(err2))
	})
}

func Test_Unwrap(t *testing.T) {
	err1 := fmt.Errorf("err1")
	err2 := errors.E("err2")
	cases := []struct {
		name     string
		err      error
		expected error
	}{
		{"nil", nil, nil},
		{"string error", err1, nil},
		{"error", err2, nil},
		{"wrapped string error", errors.E(err1), err1},
		{"wrapped error", errors.E("err", err2), err2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, errors.Unwrap(tc.err))
		})
	}
}

func Test_Trace(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.Nil(t, errors.Trace(nil))
	})

	t.Run("string error without trace", func(t *testing.T) {
		assert.Nil(t, errors.Trace(fmt.Errorf("")))
	})

	t.Run("simple error", func(t *testing.T) {
		err := errors.E("msg")
		trace := errors.Trace(err)
		assert.NotNil(t, trace)
		assert.Contains(t, fmt.Sprintf("%+v", trace[0]), "Test_Trace")
	})

	t.Run("wrapped error", func(t *testing.T) {
		err1 := findUser(1)
		err2 := errors.E("msg", err1)
		trace := errors.Trace(err2)
		assert.NotNil(t, trace)
		assert.Contains(t, fmt.Sprintf("%+v", trace[0]), "findUser")
		assert.Contains(t, fmt.Sprintf("%+v", trace[1]), "Test_Trace")
	})

	t.Run("formatting", func(t *testing.T) {
		err := findUser(1)
		trace := errors.Trace(err)

		t.Run("empty frame", func(t *testing.T) {
			trace := errors.StackTrace{errors.StackFrame(0)}
			assert.Contains(t, fmt.Sprintf("%s", trace), "unknown")
			assert.Contains(t, fmt.Sprintf("%+s", trace), "unknown")
			assert.Contains(t, fmt.Sprintf("%v", trace), "unknown")
			assert.Contains(t, fmt.Sprintf("%+v", trace), "unknown")
			assert.Contains(t, fmt.Sprintf("%#v", trace), "unknown")
			b, err := trace[0].MarshalText()
			assert.NoError(t, err, "text marshal error")
			assert.Contains(t, string(b), "unknown")
		})

		t.Run("%s", func(t *testing.T) {
			s := fmt.Sprintf("%s", trace)
			assert.True(t, strings.Contains(s, "error_test.go"))
		})

		t.Run("%+s", func(t *testing.T) {
			s := fmt.Sprintf("%+s", trace)
			assert.True(t, strings.Contains(s, "error_test.go"))
			assert.True(t, strings.Contains(s, "Test_Trace"))
		})

		t.Run("%v", func(t *testing.T) {
			s := fmt.Sprintf("%v", trace)
			fmt.Printf(s)
			assert.True(t, strings.Contains(s, "error_test.go:"))
		})

		t.Run("%+v", func(t *testing.T) {
			s := fmt.Sprintf("%+v", trace)
			assert.True(t, strings.Contains(s, "github.com/w1ck3dg0ph3r/go-errors_test.findUser"))
			assert.True(t, strings.Contains(s, "github.com/w1ck3dg0ph3r/go-errors_test.Test_Trace"))
		})

		t.Run("%#v", func(t *testing.T) {
			s := fmt.Sprintf("%#v", trace)
			assert.True(t, strings.HasPrefix(s, "[]errors.StackFrame{"))
			assert.True(t, strings.Contains(s, "error_test.go:"))
			assert.True(t, strings.HasSuffix(s, "}"))
		})

		t.Run("%n", func(t *testing.T) {
			s := fmt.Sprintf("%n", trace)
			assert.True(t, strings.Contains(s, "findUser"))
			assert.True(t, strings.Contains(s, "Test_Trace"))
		})

		t.Run("json", func(t *testing.T) {
			v := struct {
				Stack errors.StackTrace `json:"stack"`
			}{Stack: trace}
			b, err := json.Marshal(v)
			assert.Nilf(t, err, "json marshal error")
			s := string(b)
			assert.True(t, strings.Contains(s, "github.com/w1ck3dg0ph3r/go-errors_test.findUser"))
			assert.True(t, strings.Contains(s, "github.com/w1ck3dg0ph3r/go-errors_test.Test_Trace"))
			assert.True(t, strings.Contains(s, "error_test.go:"))
		})
	})
}

// findUser returns different errors based on id
func findUser(id int) error {
	const op = errors.Op("db.findUser")
	if id == 1 {
		return errors.E(op, fmt.Sprintf("user not found: %d", id), errors.Client, errors.NotFound)
	}
	if id == 2 {
		return errors.E(op, "connection failure", errors.IO, errors.Server, errors.Transient)
	}
	return nil
}

type someError struct {
	code  int
	cause error
}

func (e someError) Unwrap() error {
	return e.cause
}

func (e someError) Error() string {
	return fmt.Sprintf("Error %d", e.code)
}
