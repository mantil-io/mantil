package log

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
)

const (
	levelInfo  = "info"
	levelDebug = "debug"
	levelError = "error"
)

type backendLogLine struct {
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
	Time  time.Time `json:"time"`
}

var (
	debugLogLevelEnabled = false
)

var (
	infoLog   = color.New()
	debugLog  = color.New(color.Faint)
	noticeLog = color.New(color.FgGreen, color.Bold)
	errorLog  = color.New(color.FgRed)
	fatalLog  = color.New(color.FgRed, color.Bold)
)

var logFile *os.File

func init() {
	f, err := os.OpenFile("/tmp/mantil.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening file: %v", err)
		return
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile)
	logFile = f
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

func Printf(format string, v ...interface{}) {
	if logFile == nil {
		return
	}
	log.Output(2, fmt.Sprintf(format, v...))
}

func Info(format string, v ...interface{}) {
	infoLog.PrintlnFunc()(sprintf(format, v...))
}

func Debug(format string, v ...interface{}) {
	if !debugLogLevelEnabled {
		return
	}
	debugLog.PrintlnFunc()(sprintf(format, v...))
}

func Notice(format string, v ...interface{}) {
	noticeLog.PrintlnFunc()(sprintf(format, v...))
}

func Backend(msg string) {
	Printf("backend: %s", msg)
	l := &backendLogLine{}
	if err := json.Unmarshal([]byte(msg), &l); err != nil {
		infoLog.PrintFunc()(sprintf("λ %s", msg))
		return
	}
	if l.Level == levelDebug && !debugLogLevelEnabled {
		return
	}
	c := levelColor(l.Level)
	c.PrintlnFunc()(sprintf("λ %s", l.Msg))
}

func Error(err error) {
	Errorf(err.Error())
}

func Errorf(format string, v ...interface{}) {
	errorLog.PrintlnFunc()(sprintf(format, v...))
}

func Fatal(err error) {
	fatalLog.PrintlnFunc()(sprintf("%v", err))
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	fatalLog.PrintlnFunc()(sprintf(format, v...))
	os.Exit(1)
}

func sprintf(format string, v ...interface{}) string {
	return fmt.Sprintf(format, v...)
}

func levelColor(level string) *color.Color {
	switch level {
	case levelInfo:
		return infoLog
	case levelDebug:
		return debugLog
	case levelError:
		return errorLog
	default:
		return infoLog
	}
}

func EnableDebugLogLevel() {
	debugLogLevelEnabled = true
}

func DisableColor() {
	noticeLog = color.New(color.Bold)
	errorLog.DisableColor()
	fatalLog = color.New(color.Bold)
}
