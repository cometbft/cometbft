package log

// Logger is what any CometBFT library should take.
type Logger interface {
	Debug(msg string, keyvals ...any)
	Info(msg string, keyvals ...any)
	Error(msg string, keyvals ...any)

	With(keyvals ...any) Logger
}
