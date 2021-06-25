package errors_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/w1ck3dg0ph3r/go-errors"
)

func Test_Group(t *testing.T) {
	var err1 error = errors.E("err1")
	var err2 = fmt.Errorf("err2")

	cases := []struct {
		name     string
		errs     []error
		expected interface{}
	}{
		{"no subtasks", []error{}, nil},
		{"nil", []error{nil}, nil},
		{"one error", []error{err1}, errors.List{err1}},
		{"one error and nils", []error{nil, err1}, errors.List{err1}},
		{"multiple errors", []error{err1, err2}, errors.List{err1, err2}},
		{"multiple errors and nils", []error{err1, nil, nil, err2, nil}, errors.List{err2, err1}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := &errors.Group{}
			for _, err := range tc.errs {
				err := err
				g.Go(func() error {
					time.Sleep(time.Duration(rand.Intn(42)) * time.Millisecond)
					return err
				})
			}
			err := g.Wait()
			if list, ok := err.(errors.List); ok {
				assert.ElementsMatchf(t, tc.expected, list, "")
			} else {
				assert.Equal(t, tc.expected, err)

			}
		})
	}
}
