package util

import (
	"github.com/CodisLabs/codis/pkg/utils/bytesize"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"strings"
)

var (
	maxFileFrag       = 10000000
	maxFragSize int64 = bytesize.GB * 1
)

func SetLogLevel(level string) string {
	level = strings.ToLower(level)
	var l = log.LEVEL_INFO
	switch level {
	case "error":
		l = log.LEVEL_ERROR
	case "warn", "warning":
		l = log.LEVEL_WARN
	case "debug":
		l = log.LEVEL_DEBUG
	case "info":
		fallthrough
	default:
		level = "info"
		l = log.LEVEL_INFO
	}
	log.SetLevel(l)
	log.Infof("set log level to <%s>", level)

	return level
}

func InitLog(file string) {
	// set output log file
	if "" != file {
		f, err := log.NewRollingFile(file, maxFileFrag, maxFragSize)
		if err != nil {
			log.PanicErrorf(err, "open rolling log file failed: %s", file)
		} else {
			defer f.Close()
			log.StdLog = log.New(f, "")
		}
	}

	log.SetLevel(log.LEVEL_INFO)
	log.SetFlags(log.Flags() | log.Lshortfile)
}
