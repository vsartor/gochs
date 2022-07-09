package log

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

var (
	lvl2str = map[int]string{
		0: "<b><gr>[DEBG]<r>",
		1: "<b><bl>[INFO]<r>",
		2: "<b><ye>[WARN]<r>",
		3: "<b><re>[ERRO]<r>",
	}
	debug = false
)

func parseAnsi(s string) string {
	s = strings.ReplaceAll(s, "<b>", "\u001b[1m")
	s = strings.ReplaceAll(s, "<r>", "\u001b[0m")
	s = strings.ReplaceAll(s, "<re>", "\u001b[31m")
	s = strings.ReplaceAll(s, "<bl>", "\u001b[34m")
	s = strings.ReplaceAll(s, "<ye>", "\u001b[33m")
	s = strings.ReplaceAll(s, "<gr>", "\u001b[38;5;243m")
	return s
}

func AllowDbg() {
	debug = true
}

func Log(lvl int, msg string, args ...interface{}) {
	_, file, ln, ok := runtime.Caller(2)
	if !ok {
		file = "<unknown>.go"
		ln = 0
	} else {
		// trim filepath starting after "gochs/"
		idx := strings.Index(file, "gochs/")
		if idx == -1 {
			panic("caller should be within gochs/")
		}
		file = file[idx+len("gochs/"):]
	}

	fmtStr := fmt.Sprintf("%s <gr>(%s:%d)<r> %s\n", lvl2str[lvl], file, ln, msg)
	logMsg := fmt.Sprintf(fmtStr, args...)
	fmt.Printf(parseAnsi(logMsg))
}

func Dbg(msg string, args ...interface{}) {
	if debug {
		Log(0, msg, args...)
	}
}

func Info(msg string, args ...interface{}) {
	Log(1, msg, args...)
}

func Warn(msg string, args ...interface{}) {
	Log(2, msg, args...)
}

func Err(msg string, args ...interface{}) error {
	Log(3, msg, args...)
	return fmt.Errorf(msg, args...)
}

func Fatal(msg string, args ...interface{}) {
	Log(3, msg, args...)
	os.Exit(1)
}
