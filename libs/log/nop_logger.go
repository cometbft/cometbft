package log

type nopLogger struct{}

// Interface assertions.
var _ Logger = (*nopLogger)(nil)

// NewNopLogger returns a logger that doesn't do anything.
func NewNopLogger() Logger { return &nopLogger{} }

func (nopLogger) Error(string, ...any)  {}
func (nopLogger) Warn(string, ...any)   {}
func (nopLogger) Info(string, ...any)   {}
func (nopLogger) Debug(string, ...any)  {}
func (l *nopLogger) With(...any) Logger { return l }
func (nopLogger) Impl() any             { return nil }
