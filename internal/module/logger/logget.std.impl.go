package logger

import (
	"fmt"
	"log"
	"os"
)

func NewSTDLogger() Logger {
	return &logger{
		l: log.New(os.Stdout, "", 0),
	}
}

type logger struct {
	l *log.Logger
}

func (l *logger) Debugf(format string, args ...any) {
	l.l.Println("DEBUG:", fmt.Sprintf(format, args...))
}

func (l *logger) Infof(format string, args ...any) {
	l.l.Println("INFO:", fmt.Sprintf(format, args...))
}

func (l *logger) Errorf(format string, args ...any) {
	l.l.Println("ERROR:", fmt.Sprintf(format, args...))
}

func (l *logger) Fatalf(format string, args ...any) {
	l.l.Println("FATAL:", fmt.Sprintf(format, args...))
	os.Exit(1) //nolint:revive
}
