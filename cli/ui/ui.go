package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mantil-io/mantil/cli/log"
)

func init() {
	if noColor() {
		print := func(a ...interface{}) {
			fmt.Println(a...)
		}
		std = &Logger{
			infoLog:   print,
			debugLog:  print,
			noticeLog: print,
			errorLog: func(a ...interface{}) {
				fmt.Printf("Error: %s\n", fmt.Sprint(a...))
			},
			fatalLog: func(a ...interface{}) {
				fmt.Printf("Fatal: %s\n", fmt.Sprint(a...))
			},
		}
		return
	}
	std = &Logger{
		infoLog:   color.New().PrintlnFunc(),
		debugLog:  color.New(color.Faint).PrintlnFunc(),
		noticeLog: color.New(color.FgGreen, color.Bold).PrintlnFunc(),
		errorLog:  color.New(color.FgRed).PrintlnFunc(),
		fatalLog:  color.New(color.FgRed, color.Bold).PrintlnFunc(),
	}
}

func noColor() bool {
	for _, a := range os.Args {
		if a == "--no-color" {
			return true
		}
	}
	return false
}

// standard logger used by package exported functions
var std *Logger

type printFunc func(a ...interface{})

type Logger struct {
	infoLog   printFunc
	debugLog  printFunc
	noticeLog printFunc
	errorLog  printFunc
	fatalLog  printFunc
}

func (u *Logger) Info(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.PrintfWithCallDepth(2, "[cli.Info] %s", msg)
	u.infoLog(msg)
}

func (u *Logger) Debug(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.PrintfWithCallDepth(2, "[cli.Debug] %s", msg)
	u.debugLog(msg)
}

func (u *Logger) Notice(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.PrintfWithCallDepth(2, "[cli.Notice] %s", msg)
	u.noticeLog(msg)
}

func (u *Logger) Error(err error) {

	var ue *log.UserError
	if errors.As(err, &ue) {
		msg := ue.Message()
		log.PrintfWithCallDepth(2, "[cli.Error] %s", msg)
		u.errorLog(msg)
		return
	}
	msg := err.Error()
	log.PrintfWithCallDepth(2, "[cli.Error] %s", msg)
	u.Errorf(msg)
}

func (u *Logger) Errorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.PrintfWithCallDepth(2, "[cli.Error] %s", msg)
	u.errorLog(msg)
}

func (u *Logger) Fatal(err error) {
	msg := fmt.Sprintf("%v", err)
	log.PrintfWithCallDepth(2, "[cli.Fatal] %s", msg)
	u.fatalLog(msg)
	os.Exit(1)
}

func (u *Logger) Fatalf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.PrintfWithCallDepth(2, "[cli.Fatal] %s", msg)
	u.fatalLog(msg)
	os.Exit(1)
}

func (u *Logger) Backend(msg string) {
	l := &backendLogLine{}
	if err := json.Unmarshal([]byte(msg), &l); err != nil {
		if strings.HasSuffix(msg, "\n") {
			msg = msg[0 : len(msg)-1]
		}
		u.infoLog(fmt.Sprintf("λ %s", msg))
		return
	}
	c := u.levelColor(l.Level)
	c(fmt.Sprintf("λ %s", l.Msg))
}

const (
	levelDebug = "debug"
	levelInfo  = "info"
	levelError = "error"
)

type backendLogLine struct {
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
	Time  time.Time `json:"time"`
}

func (u *Logger) levelColor(level string) printFunc {
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

func Info(format string, v ...interface{}) {
	std.Info(format, v...)
}

func Debug(format string, v ...interface{}) {
	std.Debug(format, v...)
}

func Notice(format string, v ...interface{}) {
	std.Notice(format, v...)
}

func Error(err error) {
	std.Error(err)
}

func Errorf(format string, v ...interface{}) {
	std.Errorf(format, v...)
}

func Fatal(err error) {
	std.Fatal(err)
}

func Fatalf(format string, v ...interface{}) {
	std.Fatalf(format, v...)
}

func Backend(msg string) {
	std.Backend(msg)
}
