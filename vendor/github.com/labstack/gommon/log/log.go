package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"strconv"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/valyala/fasttemplate"

	"github.com/labstack/gommon/color"
)

type (
	Logger struct {
		prefix     string
		level      uint8
		output     io.Writer
		template   *fasttemplate.Template
		levels     []string
		color      *color.Color
		bufferPool sync.Pool
		mutex      sync.Mutex
	}
)

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
	FATAL
	OFF
)

var (
	global        = New("-")
	defaultFormat = "time=${time_rfc3339}, level=${level}, prefix=${prefix}, file=${short_file}, " +
		"line=${line}, message=${message}\n"
)

func New(prefix string) (l *Logger) {
	l = &Logger{
		level:    INFO,
		prefix:   prefix,
		template: l.newTemplate(defaultFormat),
		color:    color.New(),
		bufferPool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 256))
			},
		},
	}
	l.initLevels()
	l.SetOutput(colorable.NewColorableStdout())
	return
}

func (l *Logger) initLevels() {
	l.levels = []string{
		l.color.Blue("DEBUG"),
		l.color.Green("INFO"),
		l.color.Yellow("WARN"),
		l.color.Red("ERROR"),
		l.color.RedBg("FATAL"),
	}
}

func (l *Logger) newTemplate(format string) *fasttemplate.Template {
	return fasttemplate.New(format, "${", "}")
}

func (l *Logger) DisableColor() {
	l.color.Disable()
	l.initLevels()
}

func (l *Logger) EnableColor() {
	l.color.Enable()
	l.initLevels()
}

func (l *Logger) Prefix() string {
	return l.prefix
}

func (l *Logger) SetPrefix(p string) {
	l.prefix = p
}

func (l *Logger) Level() uint8 {
	return l.level
}

func (l *Logger) SetLevel(v uint8) {
	l.level = v
}

func (l *Logger) Output() io.Writer {
	return l.output
}

func (l *Logger) SetFormat(f string) {
	l.template = l.newTemplate(f)
}

func (l *Logger) SetOutput(w io.Writer) {
	l.output = w
	if w, ok := w.(*os.File); !ok || !isatty.IsTerminal(w.Fd()) {
		l.DisableColor()
	}
}

func (l *Logger) Print(i ...interface{}) {
	fmt.Fprintln(l.output, i...)
}

func (l *Logger) Printf(format string, args ...interface{}) {
	f := fmt.Sprintf("%s\n", format)
	fmt.Fprintf(l.output, f, args...)
}

func (l *Logger) Debug(i ...interface{}) {
	l.log(DEBUG, "", i...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(i ...interface{}) {
	l.log(INFO, "", i...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(i ...interface{}) {
	l.log(WARN, "", i...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(i ...interface{}) {
	l.log(ERROR, "", i...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) Fatal(i ...interface{}) {
	l.log(FATAL, "", i...)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
	os.Exit(1)
}

func DisableColor() {
	global.DisableColor()
}

func EnableColor() {
	global.EnableColor()
}

func Prefix() string {
	return global.Prefix()
}

func SetPrefix(p string) {
	global.SetPrefix(p)
}

func Level() uint8 {
	return global.Level()
}

func SetLevel(v uint8) {
	global.SetLevel(v)
}

func Output() io.Writer {
	return global.Output()
}

func SetOutput(w io.Writer) {
	global.SetOutput(w)
}

func SetFormat(f string) {
	global.SetFormat(f)
}

func Print(i ...interface{}) {
	global.Print(i...)
}

func Printf(format string, args ...interface{}) {
	global.Printf(format, args...)
}

func Debug(i ...interface{}) {
	global.Debug(i...)
}

func Debugf(format string, args ...interface{}) {
	global.Debugf(format, args...)
}

func Info(i ...interface{}) {
	global.Info(i...)
}

func Infof(format string, args ...interface{}) {
	global.Infof(format, args...)
}

func Warn(i ...interface{}) {
	global.Warn(i...)
}

func Warnf(format string, args ...interface{}) {
	global.Warnf(format, args...)
}

func Error(i ...interface{}) {
	global.Error(i...)
}

func Errorf(format string, args ...interface{}) {
	global.Errorf(format, args...)
}

func Fatal(i ...interface{}) {
	global.Fatal(i...)
}

func Fatalf(format string, args ...interface{}) {
	global.Fatalf(format, args...)
}

func (l *Logger) log(v uint8, format string, args ...interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	buf := l.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer l.bufferPool.Put(buf)
	_, file, line, _ := runtime.Caller(3)

	if v >= l.level {
		message := ""
		if format == "" {
			message = fmt.Sprint(args...)
		} else {
			message = fmt.Sprintf(format, args...)
		}
		if v == FATAL {
			stack := make([]byte, 4<<10)
			length := runtime.Stack(stack, true)
			message = message + "\n" + string(stack[:length])
		}
		_, err := l.template.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
			switch tag {
			case "time_rfc3339":
				return w.Write([]byte(time.Now().Format(time.RFC3339)))
			case "level":
				return w.Write([]byte(l.levels[v]))
			case "prefix":
				return w.Write([]byte(l.prefix))
			case "long_file":
				return w.Write([]byte(file))
			case "short_file":
				return w.Write([]byte(path.Base(file)))
			case "line":
				return w.Write([]byte(strconv.Itoa(line)))
			case "message":
				return w.Write([]byte(message))
			default:
				return w.Write([]byte(fmt.Sprintf("[unknown tag %s]", tag)))
			}
		})
		if err == nil {
			l.output.Write(buf.Bytes())
		}
	}
}
