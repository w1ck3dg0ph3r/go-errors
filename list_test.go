package errors_test

import (
	stderr "errors"
	"fmt"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/w1ck3dg0ph3r/go-errors"
)

func Test_Multiple(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var err error = errors.List{}
		assert.Empty(t, errors.Multiple(err))
	})

	t.Run("not a list", func(t *testing.T) {
		err := fmt.Errorf("")
		assert.Len(t, errors.Multiple(err), 1)
	})

	t.Run("one error", func(t *testing.T) {
		list := errors.List{}
		list.Add(fmt.Errorf(""))
		assert.Len(t, errors.Multiple(list), 1)
	})

	t.Run("several errors", func(t *testing.T) {
		list := errors.List{}
		list.Add(fmt.Errorf("err1"))
		list.Add(nil)
		list.Add(errors.E("err2"))
		assert.Len(t, errors.Multiple(list), 2)
	})

	t.Run("clear", func(t *testing.T) {
		list := errors.List{}
		list.Add(fmt.Errorf("err1"))
		list.Add(errors.E("err2"))
		list.Clear()
		assert.Len(t, errors.Multiple(list), 0)
	})
}

func Test_Has(t *testing.T) {
	t.Run("not a list", func(t *testing.T) {
		err1 := fmt.Errorf("")
		err2 := errors.E(errors.Server, err1)
		assert.True(t, errors.Has(err2, err1))
		assert.True(t, errors.Has(err2, errors.Server))
		assert.False(t, errors.Has(err2, errors.Client))
	})

	t.Run("empty list", func(t *testing.T) {
		var err error = errors.List{}
		assert.False(t, errors.Has(err, fs.ErrInvalid))
		assert.False(t, errors.Has(err, errors.Server))
		assert.False(t, errors.Has(err, errors.IO))
	})

	t.Run("list", func(t *testing.T) {
		list := errors.List{}
		list.Add(fmt.Errorf("error one"))
		list.Add(fs.ErrClosed)
		list.Add(someError{code: 444})
		list.Add(errors.E(errors.Op("op"), "err3", errors.Client, errors.Invalid))
		err := list.ErrOrNil()
		assert.True(t, errors.Has(err, errors.Client))
		assert.True(t, errors.Has(err, errors.Invalid))
		assert.True(t, errors.Has(err, fs.ErrClosed))
		assert.False(t, errors.Has(err, errors.Server))
		assert.False(t, errors.Has(err, errors.IO))
		assert.False(t, errors.Has(err, fs.ErrNotExist))
	})
}

func Test_HasAnyOf(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		assert.False(t, errors.HasAnyOf(nil, errors.Server, fs.ErrInvalid))
	})

	t.Run("empty", func(t *testing.T) {
		assert.False(t, errors.HasAnyOf(errors.List{}, errors.Server, fs.ErrInvalid))
	})

	t.Run("list", func(t *testing.T) {
		list := errors.List{}
		list.Add(fmt.Errorf("error one"))
		list.Add(fs.ErrClosed)
		list.Add(someError{code: 444})
		list.Add(errors.E(errors.Op("op"), "err3", errors.Client, errors.Invalid))
		err := list.ErrOrNil()
		assert.True(t, errors.HasAnyOf(err, errors.Client, errors.Server))
		assert.True(t, errors.HasAnyOf(err, errors.Invalid, errors.NotFound))
		assert.True(t, errors.HasAnyOf(err, fs.ErrNotExist, fs.ErrClosed))
		assert.False(t, errors.HasAnyOf(err, errors.Server, errors.Transient))
		assert.False(t, errors.HasAnyOf(err, errors.IO, errors.AlreadyExists))
		assert.False(t, errors.HasAnyOf(err, fs.ErrNotExist, fs.ErrInvalid))
	})
}

func Test_ListIs(t *testing.T) {
	t.Run("not a list", func(t *testing.T) {
		err1 := fmt.Errorf("")
		err2 := errors.E(errors.Server, err1)

		assert.True(t, errors.Is(err2, err1))
		assert.True(t, errors.Is(err2, errors.Server))
		assert.False(t, errors.Is(err2, errors.Client))

		assert.True(t, stderr.Is(err2, err1))
	})

	t.Run("empty list", func(t *testing.T) {
		var err error = errors.List{}
		assert.False(t, errors.Is(err, fs.ErrInvalid))
		assert.False(t, errors.Is(err, errors.Server))
		assert.False(t, errors.Is(err, errors.IO))

		assert.False(t, stderr.Is(err, fs.ErrInvalid))
	})

	t.Run("list", func(t *testing.T) {
		list := errors.List{}
		list.Add(fmt.Errorf("error one"))
		list.Add(fs.ErrClosed)
		list.Add(someError{code: 444})
		list.Add(errors.E(errors.Op("op"), "err3", errors.Client, errors.Invalid))
		err := list.ErrOrNil()

		assert.True(t, errors.Is(err, errors.Client))
		assert.True(t, errors.Is(err, errors.Invalid))
		assert.True(t, errors.Is(err, fs.ErrClosed))
		assert.False(t, errors.Is(err, errors.Server))
		assert.False(t, errors.Is(err, errors.IO))
		assert.False(t, errors.Is(err, fs.ErrNotExist))

		assert.True(t, stderr.Is(err, fs.ErrClosed))
		assert.False(t, stderr.Is(err, fs.ErrNotExist))
	})
}

func Test_ListAs(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		var err error = &errors.List{}
		var res *errors.Error
		assert.False(t, errors.As(err, &res))
		assert.False(t, stderr.As(err, &res))
	})

	t.Run("list", func(t *testing.T) {
		err1 := someError{code: 444}
		err2 := errors.E("err2")
		var err error = errors.List{err1, err2}

		var res1 someError
		var res2 *errors.Error
		var res3 otherError

		assert.True(t, errors.As(err, &res1))
		assert.Equal(t, 444, res1.code)
		assert.True(t, errors.As(err, &res2))
		assert.Equal(t, "err2", res2.Msg)
		assert.False(t, errors.As(err, &res3))

		assert.True(t, stderr.As(err, &res1))
		assert.Equal(t, 444, res1.code)
		assert.True(t, stderr.As(err, &res2))
		assert.Equal(t, "err2", res2.Msg)
		assert.False(t, stderr.As(err, &res3))
	})

	t.Run("list with wrapped error", func(t *testing.T) {
		err1 := someError{code: 444}
		err2 := errors.E("err2", err1)
		var err error = errors.List{err2}

		var res1 someError
		var res2 *errors.Error
		var res3 otherError

		assert.True(t, errors.As(err, &res1))
		assert.Equal(t, 444, res1.code)
		assert.True(t, errors.As(err, &res2))
		assert.Equal(t, "err2", res2.Msg)
		assert.False(t, errors.As(err, &res3))

		assert.True(t, stderr.As(err, &res1))
		assert.Equal(t, 444, res1.code)
		assert.True(t, stderr.As(err, &res2))
		assert.Equal(t, "err2", res2.Msg)
		assert.False(t, stderr.As(err, &res3))
	})

	t.Run("nested list", func(t *testing.T) {
		err1 := someError{code: 444}
		err2 := errors.E("err2")
		var list1 error = errors.List{err1}
		var list2 error = errors.List{err2, list1}

		var res1 someError
		var res2 *errors.Error
		var res3 otherError

		assert.True(t, errors.As(list2, &res1))
		assert.Equal(t, 444, res1.code)
		assert.True(t, errors.As(list2, &res2))
		assert.Equal(t, "err2", res2.Msg)
		assert.False(t, errors.As(list2, &res3))

		assert.True(t, stderr.As(list2, &res1))
		assert.Equal(t, 444, res1.code)
		assert.True(t, stderr.As(list2, &res2))
		assert.Equal(t, "err2", res2.Msg)
		assert.False(t, stderr.As(list2, &res3))
	})
}

func Test_List_ErrorOrNil(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		list := &errors.List{}
		assert.Nil(t, list.ErrOrNil())
	})

	t.Run("with errors", func(t *testing.T) {
		list := errors.List{}
		list.Add(fmt.Errorf(""))
		err := list.ErrOrNil()
		assert.Equal(t, list, err)
	})
}

type otherError struct{}

func (otherError) Error() string {
	return ""
}
