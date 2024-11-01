package log

import (
	"io"
	"log/slog"
)

// NewTMJSONLogger returns a Logger that encodes keyvals to the Writer as a
// single JSON object. Each log event produces no more than one call to
// w.Write. The passed Writer must be safe for concurrent use by multiple
// goroutines if the returned Logger will be used concurrently.
func NewTMJSONLogger(w io.Writer) Logger {
	logger := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return &tmLogger{logger}
}

// NewTMJSONLoggerNoTS is the same as NewTMJSONLogger, but without the
// timestamp. Used for testing purposes.
func NewTMJSONLoggerNoTS(w io.Writer) Logger {
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
