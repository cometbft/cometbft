package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/lmittmann/tint"
)

// Logger is the CometBFT logging interface.
type Logger interface {
	// Error logs a message at level ERROR.
	Error(msg string, keyvals ...any)
	// Warn logs a message at level WARN.
	Warn(msg string, keyvals ...any)
	// Info logs a message at level INFO.
	Info(msg string, keyvals ...any)
	// Debug logs a message at level DEBUG.
	Debug(msg string, keyvals ...any)

	// With returns a new contextual logger with keyvals prepended to those
	// passed to calls to Info, Warn, Debug or Error.
	With(keyvals ...any) Logger
}

type baseLogger struct {
	srcLogger *slog.Logger
}

// Interface assertions.
var _ Logger = (*baseLogger)(nil)

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
	logger := slog.New(tint.NewHandler(w, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: "2006-01-02T15:04:05.000",
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if err, ok := a.Value.Any().(error); ok {
				aErr := tint.Err(err)
				aErr.Key = a.Key
				return aErr
			}
			return a
		},
	},
	))
	return &baseLogger{slog.New(&tabHandler{h: logger.Handler()})}
}

func (l *baseLogger) Error(msg string, keyvals ...any) {
	l.srcLogger.Error(msg, keyvals...)
}

func (l *baseLogger) Warn(msg string, keyvals ...any) {
	l.srcLogger.Warn(msg, keyvals...)
}

func (l *baseLogger) Info(msg string, keyvals ...any) {
	l.srcLogger.Info(msg, keyvals...)
}

func (l *baseLogger) Debug(msg string, keyvals ...any) {
	if LogDebug {
		l.srcLogger.Debug(msg, keyvals...)
	}
}

func (l *baseLogger) With(keyvals ...any) Logger {
	return &baseLogger{l.srcLogger.With(keyvals...)}
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
	return &baseLogger{logger}
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
	return &baseLogger{logger}
}

// tabHandler is a slog.Handler that adds two tabs between the message and the attributes.
type tabHandler struct {
	h slog.Handler
}

func (th tabHandler) Handle(ctx context.Context, r slog.Record) error {
	// Format the message with some spaces between the message and the attributes.
	formattedMsg := fmt.Sprintf("%-44s", r.Message)

	// Create a new Record with the formatted message.
	record := slog.NewRecord(r.Time, r.Level, formattedMsg, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		record.Add(a)
		return true
	})
	return th.h.Handle(ctx, record)
}

func (th *tabHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return th.h.Enabled(ctx, lvl)
}

func (th *tabHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &tabHandler{h: th.h.WithAttrs(attrs)}
}

func (th *tabHandler) WithGroup(name string) slog.Handler {
	return &tabHandler{h: th.h.WithGroup(name)}
}
