package log

import (
	"io"
	"log/slog"

	"github.com/lmittmann/tint"
)

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
