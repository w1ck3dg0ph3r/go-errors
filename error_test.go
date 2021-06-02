package errors_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/w1ck3dg0ph3r/go-errors"
)

func TestConstructor(t *testing.T) {
	var err error

	err = errors.E(nil)
	assert.Nil(t, err)

	assert.PanicsWithValue(t, "bad call to E", func() {
		err = errors.E(42)
	})
}

func TestKind(t *testing.T) {
	var err error

	err = fmt.Errorf("")
	assert.Equal(t, errors.ErrorKind(0), errors.Kind(err))

	err = errors.E("kindless")
	assert.Equal(t, errors.ErrorKind(0), errors.Kind(err))

	err = errors.E(errors.Client)
	assert.Equal(t, errors.Kind(err), errors.Client)

	err = errors.E(errors.Server, errors.Transient)
	assert.True(t, errors.Is(err, errors.Server))
	assert.True(t, errors.Is(err, errors.Transient))
	assert.False(t, errors.Is(err, errors.Client))

	err = errors.E(errors.Server)
	err = errors.E("wrapped", err)
	assert.Equal(t, errors.Server, errors.Kind(err))
}

func TestWrappedKind(t *testing.T) {
	err1 := errors.E("1", errors.Server)
	err2 := errors.E(err1, "2", errors.Client)
	assert.True(t, errors.Is(err2, errors.Client))
	assert.False(t, errors.Is(err2, errors.Server))
}

func TestErrorCode(t *testing.T) {
	var err error

	err = fmt.Errorf("")
	assert.Equal(t, errors.Unexpected, errors.Code(err))

	err = errors.E()
	assert.Equal(t, errors.Unexpected, errors.Code(err))

	err = errors.E(errors.Server, errors.Transient, errors.IO)
	assert.Equal(t, errors.IO, errors.Code(err))

	err = errors.E(errors.Client, errors.NotFound)
	assert.Equal(t, errors.NotFound, errors.Code(err))

	err = errors.E(errors.NotFound)
	err = errors.E("wrapped", err)
	assert.Equal(t, errors.NotFound, errors.Code(err))
}

func TestErrorTraits(t *testing.T) {
	var err error

	err = fmt.Errorf("")
	assert.True(t, errors.Is(err, errors.Unexpected))
	assert.False(t, errors.Is(err, errors.Server))
	assert.False(t, errors.Is(err, errors.IO))

	err = errors.E("msg")
	assert.False(t, errors.Is(err, errors.Client))
	assert.False(t, errors.Is(err, errors.Server))
	assert.False(t, errors.Is(err, errors.NotFound))

	err = errors.E(errors.Server, errors.IO)
	assert.True(t, errors.IsAnyOf(err, errors.Server, errors.Transient))
	assert.True(t, errors.IsAnyOf(err, errors.Deadlock, errors.IO))
	assert.False(t, errors.IsAnyOf(err, errors.Client, errors.Invalid))
	assert.False(t, errors.IsAnyOf(err, errors.AlreadyExists, errors.NotFound))

	if err := buffUser(1, "a"); err != nil {
		assert.True(t, errors.Is(err, errors.Client))
		assert.True(t, errors.Is(err, errors.NotFound))
		assert.False(t, errors.Is(err, errors.Transient))
		assert.Equal(t, errors.ClientMsg(err), "user not found: 1")
	}
	if err := buffUser(2, "a"); err != nil {
		assert.True(t, errors.Is(err, errors.Transient))
		assert.True(t, errors.Is(err, errors.IO))
		assert.Equal(t, errors.ClientMsg(err), "")
	}
	assert.NoError(t, buffUser(3, "a"))
	if err := buffUser(3, "b"); err != nil {
		assert.True(t, errors.Is(err, errors.Client))
		assert.True(t, errors.Is(err, errors.Invalid))
		assert.False(t, errors.Is(err, errors.Transient))
		assert.Equal(t, errors.ClientMsg(err), "unknown buff: b")
	}
}

func TestErrorMessage(t *testing.T) {
	var err error

	err = fmt.Errorf("msg1")
	assert.Equal(t, "msg1", err.Error())
	assert.Equal(t, "", errors.ClientMsg(err))

	err = errors.E("msg1", errors.Client)
	assert.Equal(t, "msg1", err.Error())
	assert.Equal(t, "msg1", errors.ClientMsg(err))

	err = fmt.Errorf("msg1")
	err = errors.E("msg2", err)
	assert.Equal(t, "msg2: msg1", err.Error())
	assert.Equal(t, "", errors.ClientMsg(err))

	err = errors.E("msg1", errors.Client)
	err = errors.E("msg2", err)
	assert.Equal(t, "msg2: msg1", err.Error())
	assert.Equal(t, "msg1", errors.ClientMsg(err))
}

func TestErrorOps(t *testing.T) {
	var err error

	err = fmt.Errorf("msg1")
	assert.Empty(t, errors.Ops(err))

	err = errors.E("msg1")
	assert.Equal(t, []errors.Op{""}, errors.Ops(err))

	err = errors.E(errors.Op("op1"), "msg1")
	assert.Equal(t, []errors.Op{"op1"}, errors.Ops(err))

	err = fmt.Errorf("msg1")
	err = errors.E(errors.Op("op2"), "msg2", err)
	assert.Equal(t, []errors.Op{"op2"}, errors.Ops(err))

	err = errors.E(errors.Op("op1"), "msg1")
	err = errors.E(errors.Op("op2"), "msg2", err)
	assert.Equal(t, []errors.Op{"op2", "op1"}, errors.Ops(err))
}

func TestErrorUnwrap(t *testing.T) {
	err1 := fmt.Errorf("err1")
	err2 := errors.E("err2")
	cases := []struct {
		err      error
		expected error
	}{
		{nil, nil},
		{err1, nil},
		{err2, nil},
		{errors.E(err1), err1},
		{errors.E("err", err2), err2},
	}
	for _, tc := range cases {
		res := errors.Unwrap(tc.err)
		if res != tc.expected {
			t.Errorf("expected %v got %v", tc.expected, res)
		}
	}
}

func TestErrorAs(t *testing.T) {
	var err error
	var ok bool

	{
		err = fmt.Errorf("err1")
		var res1 error
		var res2 *errors.Error

		ok = errors.As(err, &res1)
		assert.True(t, ok)
		assert.Equal(t, "err1", res1.Error())

		ok = errors.As(err, &res2)
		assert.False(t, ok)
		assert.Nil(t, res2)
	}
	{
		err = someError{code: 1234}
		err = errors.E(errors.Op("op1"), err, errors.Client)
		var res1 *errors.Error
		var res2 someError

		ok = errors.As(err, &res1)
		assert.True(t, ok)
		assert.Equal(t, errors.Client, res1.Kind)
		assert.Equal(t, errors.Op("op1"), res1.Op)

		ok = errors.As(err, &res2)
		assert.True(t, ok)
		assert.Equal(t, 1234, res2.code)
	}
}

func TestStackTrace(t *testing.T) {
	err := findUser(1)
	trace := errors.Trace(err)

	var s string
	s = fmt.Sprintf("%+v", trace)
	assert.True(t, strings.Contains(s, "github.com/w1ck3dg0ph3r/go-errors_test.findUser"))
	assert.True(t, strings.Contains(s, "github.com/w1ck3dg0ph3r/go-errors_test.TestStackTrace"))

	s = fmt.Sprintf("%#v", trace)
	assert.True(t, strings.HasPrefix(s, "[]errors.StackFrame{"))
	assert.True(t, strings.Contains(s, "error_test.go:"))
	assert.True(t, strings.HasSuffix(s, "}"))

	s = fmt.Sprintf("%s", trace)
	assert.True(t, strings.Contains(s, "error_test.go"))

	s = fmt.Sprintf("%n", trace)
	assert.True(t, strings.Contains(s, "findUser"))
	assert.True(t, strings.Contains(s, "TestStackTrace"))

	v := struct {
		Stack errors.StackTrace `json:"stack"`
	}{Stack: trace}
	b, err := json.Marshal(v)
	assert.Nilf(t, err, "json marshal error")
	s = string(b)
	assert.True(t, strings.Contains(s, "github.com/w1ck3dg0ph3r/go-errors_test.findUser"))
	assert.True(t, strings.Contains(s, "github.com/w1ck3dg0ph3r/go-errors_test.TestStackTrace"))
	assert.True(t, strings.Contains(s, "error_test.go:"))
}

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

func buffUser(id int, buff string) error {
	const op = errors.Op("svc.buffUser")
	if err := findUser(id); err != nil {
		return errors.E(errors.Op("svc.buffUser"), err)
	}
	if buff == "b" {
		return errors.E(op, fmt.Sprintf("unknown buff: "+buff), errors.Client, errors.Invalid)
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
