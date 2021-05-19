# go-errors [![Build Status](https://travis-ci.org/w1ck3dg0ph3r/go-errors.svg?branch=master)](https://travis-ci.org/w1ck3dg0ph3r/go-errors) [![Go Report Card](https://goreportcard.com/badge/github.com/w1ck3dg0ph3r/go-errors)](https://goreportcard.com/report/github.com/w1ck3dg0ph3r/go-errors)

Package errors provides error wrapping with stack trace and human-readable operations.

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