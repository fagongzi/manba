package plugin

import (
	"github.com/fagongzi/log"
)

// LogModule log module
type LogModule struct {
}

// Info info
func (l *LogModule) Info(v ...interface{}) {
	log.Info(v...)
}

// Infof infof
func (l *LogModule) Infof(format string, v ...interface{}) {
	log.Infof(format, v...)
}

// Debug debug
func (l *LogModule) Debug(v ...interface{}) {
	log.Debug(v...)
}

// Debugf debugf
func (l *LogModule) Debugf(format string, v ...interface{}) {
	log.Debugf(format, v...)
}

// Warn warn
func (l *LogModule) Warn(v ...interface{}) {
	log.Warning(v...)
}

// Warnf warnf
func (l *LogModule) Warnf(format string, v ...interface{}) {
	log.Warningf(format, v...)
}

// Warning warning
func (l *LogModule) Warning(v ...interface{}) {
	log.Warning(v...)
}

// Warningf warningf
func (l *LogModule) Warningf(format string, v ...interface{}) {
	log.Warningf(format, v...)
}

// Error error
func (l *LogModule) Error(v ...interface{}) {
	log.Error(v...)
}

// Errorf errorf
func (l *LogModule) Errorf(format string, v ...interface{}) {
	log.Errorf(format, v...)
}

// Fatal fatal
func (l *LogModule) Fatal(v ...interface{}) {
	log.Fatal(v...)
}

// Fatalf fatalf
func (l *LogModule) Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}
