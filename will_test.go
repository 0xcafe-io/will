package will_test

import (
	"bytes"
	"errors"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/0xcafe-io/will"
)

func makeErrFunc(errToReturn error) (deferFunc func() error, called func() bool) {
	isCalled := false
	return func() error {
			isCalled = true
			return errToReturn
		}, func() bool {
			return isCalled
		}
}

func TestCaptureErr_ReportsDeferError(t *testing.T) {
	t.Cleanup(func() {})
	errDefer := errors.New("error from defer")
	funcToDefer, called := makeErrFunc(errDefer)

	funcUnderTest := func() (err error) {
		defer will.CaptureErr(funcToDefer, &err)
		if called() {
			t.Fatal("must be deferred, but run before return")
		}

		return nil
	}

	err := funcUnderTest()

	if !called() {
		t.Fatal("deferred func is not called")
	}

	if !errors.Is(err, errDefer) {
		t.Fatal("deferred error is not reported")
	}
}

func TestCaptureErr_ReportsBothErrors(t *testing.T) {
	errDefer := errors.New("error from defer")
	errReturn := errors.New("error from return")

	errFunc, called := makeErrFunc(errDefer)

	funcUnderTest := func() (err error) {
		defer will.CaptureErr(errFunc, &err)
		if called() {
			t.Fatal("must be deferred, but run before return")
		}
		return errReturn
	}

	err := funcUnderTest()

	if !called() {
		t.Fatal("cleanUp is not called")
	}

	if !errors.Is(err, errReturn) {
		t.Fatal("returned error is not reported")
	}

	if !errors.Is(err, errDefer) {
		t.Fatal("deferred error is not reported")
	}
}

func TestRecoverTo(t *testing.T) {
	t.Cleanup(func() {
		log.SetOutput(os.Stderr)
	})

	errToPanic := errors.New("be prepared")

	funcUnderTest := func() (err error) {
		defer will.RecoverTo(&err)
		panic(errToPanic)
	}

	err := funcUnderTest()
	if !errors.Is(err, errToPanic) {
		t.Fatal("panic error is not reported")
	}

	if !errors.Is(err, will.ErrPanic) {
		t.Fatal("error does not wrap will.ErrPanic")
	}
}

func TestLogErr(t *testing.T) {
	t.Cleanup(func() {
		log.SetOutput(os.Stderr)
	})

	errDefer := errors.New("error from defer")

	errFunc, called := makeErrFunc(errDefer)

	funcUnderTest := func() error {
		defer will.LogErr(errFunc)
		if called() {
			t.Fatal("must be deferred, but run before return")
		}
		return nil
	}

	buf := bytes.Buffer{}
	log.SetOutput(&buf)

	err := funcUnderTest()

	if err != nil {
		t.Fatal("expected no error")
	}

	logs := buf.String()
	if !strings.Contains(logs, "error from defer") {
		t.Fatal("expected error to be logged")
	}
}

func TestHandleErr(t *testing.T) {
	var handledErr error = nil
	errDefer := errors.New("error from defer")

	will.SetErrHandler(func(err error) {
		handledErr = err
	})

	deferFunc, _ := makeErrFunc(errDefer)

	funcUnderTest := func() error {
		defer will.HandleErr(deferFunc)
		return nil
	}

	_ = funcUnderTest()

	if !errors.Is(handledErr, errDefer) {
		t.Fatal("expected error to be handled")
	}
}
