package ui

import (
	"errors"
	"fmt"
	"os"

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
		errorLog: func(a ...interface{}) {
			color.New(color.FgRed).Print("Error: ")
			fmt.Printf("%s\n", fmt.Sprint(a...))
		},
		fatalLog: color.New(color.FgRed, color.Bold).PrintlnFunc(),
		titleLog: color.New(color.Bold).PrintFunc(),
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
	titleLog  printFunc
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

func (u *Logger) Title(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.PrintfWithCallDepth(2, "[cli.Title] %s", msg)
	u.titleLog(msg)
}

func (u *Logger) Error(err error) {
	var ue *log.UserError
	if errors.As(err, &ue) {
		for {
			msg := ue.Message()
			log.PrintfWithCallDepth(2, "[cli.Error] %s", msg)
			u.errorLog(msg)
			if !errors.As(ue.Unwrap(), &ue) {
				break
			}
		}
		return
	}
	msg := err.Error()
	log.PrintfWithCallDepth(2, "[cli.Error] %s", msg)
	u.Errorf("%s", msg)
}

func (u *Logger) Errorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.PrintfWithCallDepth(2, "[cli.Error] %s", msg)
	u.errorLog(msg)
}

func (u *Logger) ErrorLine(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.PrintfWithCallDepth(2, "[cli.ErrorLine] %s", msg)
	u.fatalLog(msg)
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

func Info(format string, v ...interface{}) {
	std.Info(format, v...)
}

func Debug(format string, v ...interface{}) {
	std.Debug(format, v...)
}

func Notice(format string, v ...interface{}) {
	std.Notice(format, v...)
}

func Title(format string, v ...interface{}) {
	std.Title(format, v...)
}

func Error(err error) {
	std.Error(err)
}

func ErrorLine(format string, v ...interface{}) {
	std.ErrorLine(format, v...)
}

func Errorf(format string, v ...interface{}) {
	std.Errorf(format, v...)
}

func Fatal(err error) {
	std.Fatal(err)
}
