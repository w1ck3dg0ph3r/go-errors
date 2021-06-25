# go-errors [![Build Status](https://travis-ci.com/w1ck3dg0ph3r/go-errors.svg?branch=master)](https://travis-ci.com/w1ck3dg0ph3r/go-errors) [![Go Report Card](https://goreportcard.com/badge/github.com/w1ck3dg0ph3r/go-errors)](https://goreportcard.com/report/github.com/w1ck3dg0ph3r/go-errors) [![codecov](https://codecov.io/gh/w1ck3dg0ph3r/go-errors/branch/master/graph/badge.svg?token=C8ZKJA2L4H)](https://codecov.io/gh/w1ck3dg0ph3r/go-errors)

Package errors provides error wrapping with stack trace and human-readable operations.

### Features

- Stacked human-readable operations associated with errors
- Customizable error kinds and codes with checking
- Error list aka multi-error
- Error group to aggregate errors from goroutines in a list
- Stack trace capture

## Install

```
go get -u github.com/w1ck3dg0ph3r/go-errors
```

## Usage

```go
package mypackage

import "github.com/w1ck3dg0ph3r/go-errors"

func MyFunc() error {
    const op = errors.Op("mypackage.MyFunc")
    
    if err := DoSmth(); err != nil {
    	if errors.IsAnyOf(err, errors.Transient, errors.Deadlock) {
    		// TODO: retry
        }
        LogError(err.Error(), errors.Ops(err), errors.Trace(err))
    	return errors.E(op, "can't do smth", err)
    }
    
    return nil
}
```