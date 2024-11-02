package log

import (
	"io"
	"log/slog"

	"github.com/lmittmann/tint"
)

// Logger is what any CometBFT library should take.
type Logger interface {
	Debug(msg string, keyvals ...any)
	Info(msg string, keyvals ...any)
	Error(msg string, keyvals ...any)

	With(keyvals ...any) Logger
}

type tmLogger struct {
	srcLogger *slog.Logger
}

// Interface assertions.
var _ Logger = (*tmLogger)(nil)

// NewLogger returns a logger that encodes msg and keyvals to the Writer
// using slog as an underlying logger and our custom formatter. Note that
// underlying logger could be swapped with something else.
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

// Info logs a message at level Info.
func (l *tmLogger) Info(msg string, keyvals ...any) {
	l.srcLogger.Info(msg, keyvals...)
}

// Debug logs a message at level Debug.
func (l *tmLogger) Debug(msg string, keyvals ...any) {
	if LogDebug {
		l.srcLogger.Debug(msg, keyvals...)
	}
}

// Error logs a message at level Error.
func (l *tmLogger) Error(msg string, keyvals ...any) {
	l.srcLogger.Error(msg, keyvals...)
}

// With returns a new contextual logger with keyvals prepended to those passed
// to calls to Info, Debug or Error.
func (l *tmLogger) With(keyvals ...any) Logger {
	return &tmLogger{l.srcLogger.With(keyvals...)}
}

// NewJSONLogger returns a Logger that encodes keyvals to the Writer as a
// single JSON object. Each log event produces no more than one call to
// w.Write. The passed Writer must be safe for concurrent use by multiple
// goroutines if the returned Logger will be used concurrently.
func NewJSONLogger(w io.Writer) Logger {
	logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return &tmLogger{logger}
}

// NewJSONLoggerNoTS is the same as NewTMJSONLogger, but without the
// timestamp. Used for testing purposes.
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
