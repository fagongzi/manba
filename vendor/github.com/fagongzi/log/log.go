package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	// Ldate flags
	Ldate = log.Ldate
	// Llongfile flags
	Llongfile = log.Llongfile
	// Lmicroseconds flags
	Lmicroseconds = log.Lmicroseconds
	// Lshortfile flags
	Lshortfile = log.Lshortfile
	// LstdFlags flags
	LstdFlags = log.LstdFlags
	// Ltime flags
	Ltime = log.Ltime
)

type (
	// Level log level
	Level int
	// Type log type
	Type int
)

const (
	fatalLevel = Type(0x1)
	errorLevel = Type(0x2)
	warnLevel  = Type(0x4)
	infoLevel  = Type(0x8)
	debugLevel = Type(0x10)
)

const (
	// LogNone log nothing
	LogNone = Level(0x0)
	// LogFatal log fatal
	LogFatal = LogNone | Level(fatalLevel)
	// LogError log error & fatal
	LogError = LogFatal | Level(errorLevel)
	// LogWarn log warn & error & fatal
	LogWarn = LogError | Level(warnLevel)
	// LogInfo log info & warn & error & fatal
	LogInfo = LogWarn | Level(infoLevel)
	// LogDebug log debug & info & warn & error & fatal
	LogDebug = LogInfo | Level(debugLevel)
	// LogAll log all
	LogAll = LogDebug
)

const (
	formatTimeDay  string = "20060102"
	formatTimeHour string = "2006010215"
)

var defaultLog = new()

func init() {
	SetFlags(Ldate | Ltime | Lmicroseconds | Lmicroseconds)
	SetHighlighting(runtime.GOOS != "windows")
}

// Logger get default Logger
func Logger() *log.Logger {
	return defaultLog._log
}

// FatalEnabled fatal enabled
func FatalEnabled() bool {
	return defaultLog.isLevelEnabled(fatalLevel)
}

// ErrorEnabled error enabled
func ErrorEnabled() bool {
	return defaultLog.isLevelEnabled(errorLevel)
}

// WarnEnabled warn enabled
func WarnEnabled() bool {
	return defaultLog.isLevelEnabled(warnLevel)
}

// InfoEnabled info enabled
func InfoEnabled() bool {
	return defaultLog.isLevelEnabled(infoLevel)
}

// DebugEnabled debug enabled
func DebugEnabled() bool {
	return defaultLog.isLevelEnabled(debugLevel)
}

// SetLevel set current log level
func SetLevel(level Level) {
	defaultLog.SetLevel(level)
}

// GetLogLevel get current log level
func GetLogLevel() Level {
	return defaultLog.level
}

// SetOutput set log file use a writer
func SetOutput(out io.Writer) {
	defaultLog.SetOutput(out)
}

// SetOutputByName set log file use file name
func SetOutputByName(path string) error {
	return defaultLog.SetOutputByName(path)
}

// SetFlags set log flags
func SetFlags(flags int) {
	defaultLog._log.SetFlags(flags)
}

// Info info
func Info(v ...interface{}) {
	defaultLog.Info(v...)
}

// Infof infof
func Infof(format string, v ...interface{}) {
	defaultLog.Infof(format, v...)
}

// Debug debug
func Debug(v ...interface{}) {
	defaultLog.Debug(v...)
}

// Debugf debugf
func Debugf(format string, v ...interface{}) {
	defaultLog.Debugf(format, v...)
}

// Warn warn
func Warn(v ...interface{}) {
	defaultLog.Warning(v...)
}

// Warnf warnf
func Warnf(format string, v ...interface{}) {
	defaultLog.Warningf(format, v...)
}

// Warning warning
func Warning(v ...interface{}) {
	defaultLog.Warning(v...)
}

// Warningf warningf
func Warningf(format string, v ...interface{}) {
	defaultLog.Warningf(format, v...)
}

// Error error
func Error(v ...interface{}) {
	defaultLog.Error(v...)
}

// Errorf errorf
func Errorf(format string, v ...interface{}) {
	defaultLog.Errorf(format, v...)
}

// Fatal fatal
func Fatal(v ...interface{}) {
	defaultLog.Fatal(v...)
}

// Fatalf fatalf
func Fatalf(format string, v ...interface{}) {
	defaultLog.Fatalf(format, v...)
}

// SetLevelByString set log by string level
func SetLevelByString(level string) {
	defaultLog.SetLevelByString(level)
}

// SetHighlighting set highlighting log
func SetHighlighting(highlighting bool) {
	defaultLog.SetHighlighting(highlighting)
}

// SetRotateByDay set default log rotate by day
func SetRotateByDay() {
	defaultLog.SetRotateByDay()
}

// SetRotateByHour set default log rotate by hour
func SetRotateByHour() {
	defaultLog.SetRotateByHour()
}

type logger struct {
	_log         *log.Logger
	level        Level
	highlighting bool

	dailyRolling bool
	hourRolling  bool

	fileName  string
	logSuffix string
	fd        *os.File

	lock sync.Mutex
}

func (l *logger) SetHighlighting(highlighting bool) {
	l.highlighting = highlighting
}

func (l *logger) SetLevel(level Level) {
	l.level = level
}

func (l *logger) SetLevelByString(level string) {
	l.level = stringToLogLevel(level)
}

func (l *logger) SetRotateByDay() {
	l.dailyRolling = true
	l.logSuffix = genDayTime(time.Now())
}

func (l *logger) SetRotateByHour() {
	l.hourRolling = true
	l.logSuffix = genHourTime(time.Now())
}

func (l *logger) rotate() error {
	l.lock.Lock()
	defer l.lock.Unlock()

	var suffix string
	if l.dailyRolling {
		suffix = genDayTime(time.Now())
	} else if l.hourRolling {
		suffix = genHourTime(time.Now())
	} else {
		return nil
	}

	// Notice: if suffix is not equal to l.LogSuffix, then rotate
	if suffix != l.logSuffix {
		err := l.doRotate(suffix)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *logger) doRotate(suffix string) error {
	// Notice: Not check error, is this ok?
	l.fd.Close()

	lastFileName := l.fileName + "." + l.logSuffix
	err := os.Rename(l.fileName, lastFileName)
	if err != nil {
		return err
	}

	err = l.SetOutputByName(l.fileName)
	if err != nil {
		return err
	}

	l.logSuffix = suffix

	return nil
}

func (l *logger) SetOutput(out io.Writer) {
	l._log = log.New(out, l._log.Prefix(), l._log.Flags())
}

func (l *logger) SetOutputByName(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}

	l.SetOutput(f)

	l.fileName = path
	l.fd = f

	return err
}

func (l *logger) log(t Type, v ...interface{}) {
	if l.level|Level(t) != l.level {
		return
	}

	err := l.rotate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return
	}

	v1 := make([]interface{}, len(v)+2)
	logStr, logColor := logTypeToString(t)
	if l.highlighting {
		v1[0] = "\033" + logColor + "m[" + logStr + "]"
		copy(v1[1:], v)
		v1[len(v)+1] = "\033[0m"
	} else {
		v1[0] = "[" + logStr + "]"
		copy(v1[1:], v)
		v1[len(v)+1] = ""
	}

	s := fmt.Sprintln(v1...)
	l._log.Output(4, s)
}

func (l *logger) logf(t Type, format string, v ...interface{}) {
	if l.level|Level(t) != l.level {
		return
	}

	err := l.rotate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		return
	}

	logStr, logColor := logTypeToString(t)
	var s string
	if l.highlighting {
		s = "\033" + logColor + "m[" + logStr + "] " + fmt.Sprintf(format, v...) + "\033[0m"
	} else {
		s = "[" + logStr + "] " + fmt.Sprintf(format, v...)
	}
	l._log.Output(4, s)
}

// FatalEnabled fatal enabled
func (l *logger) FatalEnabled() bool {
	return l.isLevelEnabled(fatalLevel)
}

// ErrorEnabled error enabled
func (l *logger) ErrorEnabled() bool {
	return l.isLevelEnabled(errorLevel)
}

// WarnEnabled warn enabled
func (l *logger) WarnEnabled() bool {
	return l.isLevelEnabled(warnLevel)
}

// InfoEnabled info enabled
func (l *logger) InfoEnabled() bool {
	return l.isLevelEnabled(infoLevel)
}

// DebugEnabled debug enabled
func (l *logger) DebugEnabled() bool {
	return l.isLevelEnabled(debugLevel)
}

func (l *logger) isLevelEnabled(target Type) bool {
	return int(l.level)&int(target) != 0
}

func (l *logger) Fatal(v ...interface{}) {
	l.log(fatalLevel, v...)
	os.Exit(-1)
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	l.logf(fatalLevel, format, v...)
	os.Exit(-1)
}

func (l *logger) Error(v ...interface{}) {
	l.log(errorLevel, v...)
}

func (l *logger) Errorf(format string, v ...interface{}) {
	l.logf(errorLevel, format, v...)
}

func (l *logger) Warning(v ...interface{}) {
	l.log(warnLevel, v...)
}

func (l *logger) Warningf(format string, v ...interface{}) {
	l.logf(warnLevel, format, v...)
}

func (l *logger) Debug(v ...interface{}) {
	l.log(debugLevel, v...)
}

func (l *logger) Debugf(format string, v ...interface{}) {
	l.logf(debugLevel, format, v...)
}

func (l *logger) Info(v ...interface{}) {
	l.log(infoLevel, v...)
}

func (l *logger) Infof(format string, v ...interface{}) {
	l.logf(infoLevel, format, v...)
}

func stringToLogLevel(level string) Level {
	switch level {
	case "fatal":
		return LogFatal
	case "error":
		return LogError
	case "warn":
		return LogWarn
	case "warning":
		return LogWarn
	case "debug":
		return LogDebug
	case "info":
		return LogInfo
	}
	return LogAll
}

func logTypeToString(t Type) (string, string) {
	switch t {
	case fatalLevel:
		return "fatal", "[0;31"
	case errorLevel:
		return "error", "[0;31"
	case warnLevel:
		return "warning", "[0;33"
	case debugLevel:
		return "debug", "[0;36"
	case infoLevel:
		return "info", "[0;37"
	}
	return "unknown", "[0;37"
}

func genDayTime(t time.Time) string {
	return t.Format(formatTimeDay)
}

func genHourTime(t time.Time) string {
	return t.Format(formatTimeHour)
}

func new() *logger {
	return newLogger(os.Stderr, "")
}

func newLogger(w io.Writer, prefix string) *logger {
	return &logger{_log: log.New(w, prefix, LstdFlags), level: LogAll, highlighting: true}
}
