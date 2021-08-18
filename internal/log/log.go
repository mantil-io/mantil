package log

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	debugLogLevelEnabled = false
)

var (
	infoLog    = color.New()
	debugLog   = color.New(color.Faint)
	noticeLog  = color.New(color.FgGreen, color.Bold)
	backendLog = color.New(color.FgHiBlue)
	errorLog   = color.New(color.FgRed)
	fatalLog   = color.New(color.FgRed, color.Bold)
)

func Info(msg string, v ...interface{}) {
	infoLog.PrintlnFunc()(sprintf(msg, v...))
}

func Debug(msg string, v ...interface{}) {
	if !debugLogLevelEnabled {
		return
	}
	debugLog.PrintlnFunc()(sprintf(msg, v...))
}

func Notice(msg string, v ...interface{}) {
	noticeLog.PrintlnFunc()(sprintf(msg, v...))
}

func Backend(msg string, v ...interface{}) {
	backendLog.PrintFunc()(sprintf(msg, v...))
}

func Error(err error) {
	errorLog.PrintlnFunc()(sprintf("%v", err))
}

func Errorf(msg string, v ...interface{}) {
	errorLog.PrintlnFunc()(sprintf(msg, v...))
}

func Fatal(err error) {
	fatalLog.PrintlnFunc()(sprintf("%v", err))
}

func Fatalf(msg string, v ...interface{}) {
	fatalLog.PrintlnFunc()(sprintf(msg, v...))
}

func EnableDebugLogLevel() {
	debugLogLevelEnabled = true
}

func sprintf(msg string, v ...interface{}) string {
	return fmt.Sprintf(msg, v...)
}

func DisableColor() {
	noticeLog = color.New(color.Bold)
	backendLog = color.New(color.Underline)
	errorLog.DisableColor()
	fatalLog = color.New(color.Bold)
}
