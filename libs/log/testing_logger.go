package log

import (
	"io"
	"os"
	"testing"

	"github.com/go-kit/log/term"
)

// reuse the same logger across all tests.
var _testingLogger Logger

// TestingLogger returns a TMLogger which writes to STDOUT if testing being run
// with the verbose (-v) flag, NopLogger otherwise.
//
// Note that the call to TestingLogger() must be made
// inside a test (not in the init func) because
// verbose flag only set at the time of testing.
func TestingLogger() Logger {
	return TestingLoggerWithOutput(os.Stdout)
}

// TestingLoggerWithOutput returns a TMLogger which writes to (w io.Writer) if testing being run
// with the verbose (-v) flag, NopLogger otherwise.
//
// Note that the call to TestingLoggerWithOutput(w io.Writer) must be made
// inside a test (not in the init func) because
// verbose flag only set at the time of testing.
func TestingLoggerWithOutput(w io.Writer) Logger {
	if _testingLogger != nil {
		return _testingLogger
	}

	if testing.Verbose() {
		_testingLogger = NewTMLogger(NewSyncWriter(w))
	} else {
		_testingLogger = NewNopLogger()
	}

	return _testingLogger
}

// TestingLoggerWithColorFn allow you to provide your own color function. See
// TestingLogger for documentation.
func TestingLoggerWithColorFn(colorFn func(keyvals ...any) term.FgBgColor) Logger {
	if _testingLogger != nil {
		return _testingLogger
	}

	if testing.Verbose() {
		_testingLogger = NewTMLoggerWithColorFn(NewSyncWriter(os.Stdout), colorFn)
	} else {
		_testingLogger = NewNopLogger()
	}

	return _testingLogger
}
