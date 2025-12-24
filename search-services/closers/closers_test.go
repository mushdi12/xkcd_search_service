package closers

import (
	"errors"
	"log/slog"
	"os"
	"testing"
)

type fakeCloser struct {
	closed bool
	err    error
}

func (f *fakeCloser) Close() error {
	f.closed = true
	return f.err
}

func TestCloseOrLog(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ok := &fakeCloser{}
	bad := &fakeCloser{err: errors.New("close error")}

	CloseOrLog(logger, ok, bad)

	if !ok.closed || !bad.closed {
		t.Fatalf("expected both closers to be closed")
	}
}

func TestCloseOrPanic(t *testing.T) {
	ok := &fakeCloser{}
	CloseOrPanic(ok)
	if !ok.closed {
		t.Fatalf("expected closer to be closed")
	}

	bad := &fakeCloser{err: errors.New("err")}
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()
	CloseOrPanic(bad)
}


