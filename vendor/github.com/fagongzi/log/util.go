package log

import (
	"flag"
)

var (
	crashLog = flag.String("crash", "./crash.log", "The crash log file.")
	logFile  = flag.String("log-file", "", "The external log file. Default log to console.")
	logLevel = flag.String("log-level", "info", "The log level, default is info")
)

// Cfg is the log cfg
type Cfg struct {
	LogLevel string
	LogFile  string
}

// InitLog init log
func InitLog() {
	if !flag.Parsed() {
		flag.Parse()
	}

	SetHighlighting(false)
	SetLevelByString(*logLevel)
	if "" != *logFile {
		SetRotateByHour()
		SetOutputByName(*logFile)
		CrashLog(*crashLog)
	}

	if !DebugEnabled() {
		SetFlags(Ldate | Ltime | Lmicroseconds)
	}
}
