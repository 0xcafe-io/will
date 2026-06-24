package will

import (
	"errors"
	"fmt"
	"log"
)

func defaultErrorHandlerFunc(err error) {
	log.Println(err.Error())
}

var errorHandlerFunc = defaultErrorHandlerFunc

// ErrPanic is the error returned by RecoverTo when a panic occurs.
var ErrPanic = errors.New("error from panic")

// LogErr logs the error returned by f().
func LogErr(f func() error) {
	if err := f(); err != nil {
		log.Println(err.Error())
	}
}

// HandleErr calls a handler set by SetErrHandler with the error returned by f().
// Default handler just logs the error.
func HandleErr(f func() error) {
	if err := f(); err != nil && errorHandlerFunc != nil {
		errorHandlerFunc(err)
	}
}

// SetErrHandler registers handler for HandleErr.
// err is always non-nil.
func SetErrHandler(h func(err error)) {
	errorHandlerFunc = h
}

// CaptureErr appends the returned error by f() into errPtr using errors.Join.
// if f() returns nil, errPtr is not modified.
// if errPtr is nil, result of f() is ignored – equivalent to will.IgnoreErr(f).
func CaptureErr(f func() error, errPtr *error) {
	err := f()
	if err == nil || errPtr == nil {
		return
	}
	*errPtr = errors.Join(*errPtr, err)
}

// IgnoreErr ignores the returned error by f().
// The goal is to make the intention explicit, reveal the signature of f, and/or satisfy the linters. Wink-wink.
func IgnoreErr(f func() error) {
	_ = f()
}

// RecoverTo recovers from potential panic and appends its value into errPtr using errors.Join.
// To check whether panic happened, use errors.Is(err, will.ErrPanic).
// If errPtr is nil, panics will not be recovered (arguably it's better than silently ignoring them).
func RecoverTo(errPtr *error) {
	if errPtr == nil {
		// panicking here would be a cruel joke
		return
	}
	if r := recover(); r != nil {
		var err error
		switch v := r.(type) {
		case error:
			err = v
		default:
			err = fmt.Errorf("%v", v)
		}

		*errPtr = errors.Join(*errPtr, err, ErrPanic)
	}
}
