package logger

type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
}
