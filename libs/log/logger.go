package log

import (
	"io"
	"log/slog"

	"github.com/lmittmann/tint"
)

// Logger is the CometBFT logging interface.
type Logger interface {
	// Error logs a message at level ERROR.
	Error(msg string, keyvals ...any)
	// Info logs a message at level INFO.
	Info(msg string, keyvals ...any)
	// Warn logs a message at level WARN.
	Warn(msg string, keyvals ...any)
	// Debug logs a message at level DEBUG.
	Debug(msg string, keyvals ...any)

	// With returns a new contextual logger with keyvals prepended to those
	// passed to calls to Info, Warn, Debug or Error.
	With(keyvals ...any) Logger

	// Impl returns the underlying logger implementation.
	// It is used to access the full functionalities of the underlying logger.
	// Advanced users can type cast the returned value to the actual logger.
	Impl() any
}

type tmLogger struct {
	srcLogger *slog.Logger
}

// Interface assertions.
var _ Logger = (*tmLogger)(nil)

// NewLogger returns a logger that writes msg and keyvals to w using slog as an
// underlying logger.
//
// github.com/lmittmann/tint library is used to colorize the output.
//
// NOTE:
//   - the underlying logger could be swapped with something else in the future
//   - w must be safe for concurrent use by multiple goroutines if the returned
//     Logger will be used concurrently.
func NewLogger(w io.Writer) Logger {
	return &tmLogger{slog.New(tint.NewHandler(w, &tint.Options{
		Level: slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if err, ok := a.Value.Any().(error); ok {
				aErr := tint.Err(err)
				aErr.Key = a.Key
				return aErr
			}
			return a
		},
	},
	))}
}

func (l *tmLogger) Error(msg string, keyvals ...any) {
	l.srcLogger.Error(msg, keyvals...)
}

func (l *tmLogger) Info(msg string, keyvals ...any) {
	l.srcLogger.Info(msg, keyvals...)
}

func (l *tmLogger) Warn(msg string, keyvals ...any) {
	l.srcLogger.Warn(msg, keyvals...)
}

func (l *tmLogger) Debug(msg string, keyvals ...any) {
	if LogDebug {
		l.srcLogger.Debug(msg, keyvals...)
	}
}

func (l *tmLogger) With(keyvals ...any) Logger {
	return &tmLogger{l.srcLogger.With(keyvals...)}
}

// Impl returns the slog.Logger.
func (l *tmLogger) Impl() any {
	return l.srcLogger
}

// NewJSONLogger returns a Logger that writes msg and keyvals to w as using
// slog (slog.NewJSONHandler).
//
// NOTE:
//   - the underlying logger could be swapped with something else in the future
//   - w must be safe for concurrent use by multiple goroutines if the returned
//     Logger will be used concurrently.
func NewJSONLogger(w io.Writer) Logger {
	logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return &tmLogger{logger}
}

// NewJSONLoggerNoTS is the same as NewJSONLogger, but without the timestamp.
// Used for testing purposes.
func NewJSONLoggerNoTS(w io.Writer) Logger {
	logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			// Remove time from the output for predictable test output.
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}

			return a
		},
	}))
	return &tmLogger{logger}
}
