package log

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

func init() {
	if noColor() {
		print := func(a ...interface{}) {
			fmt.Println(a...)
		}
		UI = &UILogger{
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
	UI = &UILogger{
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

var UI *UILogger

type printFunc func(a ...interface{})

type UILogger struct {
	infoLog   printFunc
	debugLog  printFunc
	noticeLog printFunc
	errorLog  printFunc
	fatalLog  printFunc
}

func (u *UILogger) Info(format string, v ...interface{}) {
	u.infoLog(fmt.Sprintf(format, v...))
}

func (u *UILogger) Debug(format string, v ...interface{}) {
	u.debugLog(fmt.Sprintf(format, v...))
}

func (u *UILogger) Notice(format string, v ...interface{}) {
	u.noticeLog(fmt.Sprintf(format, v...))
}

func (u *UILogger) Error(err error) {
	var ue *UserError
	if errors.As(err, &ue) {
		u.errorLog(ue.Message())
		return
	}
	u.Errorf(err.Error())
}

func (u *UILogger) Errorf(format string, v ...interface{}) {
	u.errorLog(fmt.Sprintf(format, v...))
}

func (u *UILogger) Fatal(err error) {
	u.fatalLog(fmt.Sprintf("%v", err))
	os.Exit(1)
}

func (u *UILogger) Fatalf(format string, v ...interface{}) {
	u.fatalLog(fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (u *UILogger) Backend(msg string) {
	l := &backendLogLine{}
	if err := json.Unmarshal([]byte(msg), &l); err != nil {
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

func (u *UILogger) levelColor(level string) printFunc {
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
