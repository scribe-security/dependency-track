/*
Package log contains the singleton object and helper functions for facilitating logging within the syft library.
*/
package client

import "fmt"

const (
	DEFAULT_APP_NAME = "default"
)

type nopLogger struct{}

func (l *nopLogger) Errorf(format string, args ...interface{}) {}
func (l *nopLogger) Error(args ...interface{})                 {}
func (l *nopLogger) Warnf(format string, args ...interface{})  {}
func (l *nopLogger) Warn(args ...interface{})                  {}
func (l *nopLogger) Infof(format string, args ...interface{})  {}
func (l *nopLogger) Info(args ...interface{})                  {}
func (l *nopLogger) Debugf(format string, args ...interface{}) {}
func (l *nopLogger) Debug(args ...interface{})                 {}

type Logger interface {
	Errorf(format string, args ...interface{})
	Error(args ...interface{})
	Warnf(format string, args ...interface{})
	Warn(args ...interface{})
	Infof(format string, args ...interface{})
	Info(args ...interface{})
	Debugf(format string, args ...interface{})
	Debug(args ...interface{})
}

// Log is the singleton used to facilitate logging internally within syft
var (
	Log        Logger = &nopLogger{}
	APP_NAME          = DEFAULT_APP_NAME
	LOG_FORMAT        = "[" + DEFAULT_APP_NAME + "]" + " %s"
)

func SetAppName(name string) {
	APP_NAME = name
	LOG_FORMAT = "[" + name + "]" + " %s"
}

// Errorf takes a formatted template string and template arguments for the error logging level.
func Errorf(format string, args ...interface{}) {
	format = fmt.Sprintf(LOG_FORMAT, format)
	Log.Errorf(format, args...)
}

// Error logs the given arguments at the error logging level.
func Error(args ...interface{}) {
	args = append([]interface{}{APP_NAME}, args...)
	Log.Error(args...)
}

// Warnf takes a formatted template string and template arguments for the warning logging level.
func Warnf(format string, args ...interface{}) {
	format = fmt.Sprintf(LOG_FORMAT, format)
	Log.Warnf(format, args...)
}

// Warn logs the given arguments at the warning logging level.
func Warn(args ...interface{}) {
	args = append([]interface{}{APP_NAME}, args...)
	Log.Warn(args...)
}

// Infof takes a formatted template string and template arguments for the info logging level.
func Infof(format string, args ...interface{}) {
	format = fmt.Sprintf(LOG_FORMAT, format)
	Log.Infof(format, args...)
}

// Info logs the given arguments at the info logging level.
func Info(args ...interface{}) {
	args = append([]interface{}{APP_NAME}, args...)
	Log.Info(args...)
}

// Debugf takes a formatted template string and template arguments for the debug logging level.
func Debugf(format string, args ...interface{}) {
	format = fmt.Sprintf(LOG_FORMAT, format)
	Log.Debugf(format, args...)
}

// Debug logs the given arguments at the debug logging level.
func Debug(args ...interface{}) {
	args = append([]interface{}{APP_NAME}, args...)
	Log.Debug(args...)
}
