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

var logFile *os.File

func init() {
	UI = &UILogger{
		infoLog:   color.New(),
		debugLog:  color.New(color.Faint),
		noticeLog: color.New(color.FgGreen, color.Bold),
		errorLog:  color.New(color.FgRed),
		fatalLog:  color.New(color.FgRed, color.Bold),
	}
	openLogFile()
}

func openLogFile() {
	f, err := os.OpenFile("/tmp/mantil.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening log file - %v", err)
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

var UI *UILogger

type UILogger struct {
	infoLog   *color.Color
	debugLog  *color.Color
	noticeLog *color.Color
	errorLog  *color.Color
	fatalLog  *color.Color
}

func (u *UILogger) Info(format string, v ...interface{}) {
	u.infoLog.PrintlnFunc()(sprintf(format, v...))
}

func (u *UILogger) Debug(format string, v ...interface{}) {
	if !debugLogLevelEnabled {
		return
	}
	u.debugLog.PrintlnFunc()(sprintf(format, v...))
}

func (u *UILogger) Notice(format string, v ...interface{}) {
	u.noticeLog.PrintlnFunc()(sprintf(format, v...))
}

func (u *UILogger) Backend(msg string) {
	Printf("backend: %s", msg)
	l := &backendLogLine{}
	if err := json.Unmarshal([]byte(msg), &l); err != nil {
		u.infoLog.PrintFunc()(sprintf("λ %s", msg))
		return
	}
	if l.Level == levelDebug && !debugLogLevelEnabled {
		return
	}
	c := u.levelColor(l.Level)
	c.PrintlnFunc()(sprintf("λ %s", l.Msg))
}

func (u *UILogger) Error(err error) {
	u.Errorf(err.Error())
}

func (u *UILogger) Errorf(format string, v ...interface{}) {
	u.errorLog.PrintlnFunc()(sprintf(format, v...))
}

func (u *UILogger) Fatal(err error) {
	u.fatalLog.PrintlnFunc()(sprintf("%v", err))
	os.Exit(1)
}

func (u *UILogger) Fatalf(format string, v ...interface{}) {
	u.fatalLog.PrintlnFunc()(sprintf(format, v...))
	os.Exit(1)
}

func sprintf(format string, v ...interface{}) string {
	return fmt.Sprintf(format, v...)
}

func (u *UILogger) levelColor(level string) *color.Color {
	switch level {
	case levelInfo:
		return u.infoLog
	case levelDebug:
		return u.debugLog
	case levelError:
		return u.errorLog
	default:
		return u.infoLog
	}
}

func EnableDebugLogLevel() {
	debugLogLevelEnabled = true
}

func (u *UILogger) DisableColor() {
	u.noticeLog = color.New(color.Bold)
	u.errorLog.DisableColor()
	u.fatalLog = color.New(color.Bold)
}
