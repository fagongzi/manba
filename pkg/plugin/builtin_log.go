package plugin

import (
	"github.com/fagongzi/log"
)

// Log builtin log
type Log struct {
}

// Info info
func (l *Log) Info(v ...interface{}) {
	log.Info(v...)
}

// Infof infof
func (l *Log) Infof(format string, v ...interface{}) {
	log.Infof(format, v...)
}

// Debug debug
func (l *Log) Debug(v ...interface{}) {
	log.Debug(v...)
}

// Debugf debugf
func (l *Log) Debugf(format string, v ...interface{}) {
	log.Debugf(format, v...)
}

// Warn warn
func (l *Log) Warn(v ...interface{}) {
	log.Warning(v...)
}

// Warnf warnf
func (l *Log) Warnf(format string, v ...interface{}) {
	log.Warningf(format, v...)
}

// Warning warning
func (l *Log) Warning(v ...interface{}) {
	log.Warning(v...)
}

// Warningf warningf
func (l *Log) Warningf(format string, v ...interface{}) {
	log.Warningf(format, v...)
}

// Error error
func (l *Log) Error(v ...interface{}) {
	log.Error(v...)
}

// Errorf errorf
func (l *Log) Errorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}

// Fatal fatal
func (l *Log) Fatal(v ...interface{}) {
	log.Fatal(v...)
}

// Fatalf fatalf
func (l *Log) Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}
