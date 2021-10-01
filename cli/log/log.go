package log

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
)

var (
	debugLogLevelEnabled = false
)

func EnableDebugLogLevel() {
	debugLogLevelEnabled = true
}

func (u *UILogger) DisableColor() {
	u.noticeLog = color.New(color.Bold)
	u.errorLog.DisableColor()
	u.fatalLog = color.New(color.Bold)
}

var (
	logFile *os.File
	logs    *log.Logger
	errs    *log.Logger
)

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
	fn := fmt.Sprintf("/tmp/mantil-%s.log", time.Now().Format("2006-01-02"))
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error opening log file - %v", err)
		return
	}
	logs = log.New(f, "", log.LstdFlags|log.Lmicroseconds|log.Llongfile)
	errs = log.New(f, "[ERROR] ", log.LstdFlags|log.Lmicroseconds|log.Llongfile|log.Lmsgprefix)
	logFile = f
}

func Close() {
	if logFile != nil {
		fmt.Fprintf(logFile, "\n")
		logFile.Close()
	}
}

func Printf(format string, v ...interface{}) {
	if logFile == nil {
		return
	}
	logs.Output(2, fmt.Sprintf(format, v...))
}

func Errorf(format string, v ...interface{}) {
	if logFile == nil {
		return
	}
	errs.Output(2, fmt.Sprintf(format, v...))
}

func Error(err error) {
	if logFile == nil {
		return
	}
	errs.Output(2, err.Error())
	printStack(logFile, err)
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
	u.infoLog.PrintlnFunc()(fmt.Sprintf(format, v...))
}

func (u *UILogger) Debug(format string, v ...interface{}) {
	if !debugLogLevelEnabled {
		return
	}
	u.debugLog.PrintlnFunc()(fmt.Sprintf(format, v...))
}

func (u *UILogger) Notice(format string, v ...interface{}) {
	u.noticeLog.PrintlnFunc()(fmt.Sprintf(format, v...))
}

func (u *UILogger) Error(err error) {
	var ue *UserError
	if errors.As(err, &ue) {
		u.errorLog.PrintlnFunc()(ue.Message())
		return
	}
	u.Errorf(err.Error())
}

func (u *UILogger) Errorf(format string, v ...interface{}) {
	u.errorLog.PrintlnFunc()(fmt.Sprintf(format, v...))
}

func (u *UILogger) Fatal(err error) {
	u.fatalLog.PrintlnFunc()(fmt.Sprintf("%v", err))
	os.Exit(1)
}

func (u *UILogger) Fatalf(format string, v ...interface{}) {
	u.fatalLog.PrintlnFunc()(fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (u *UILogger) Backend(msg string) {
	l := &backendLogLine{}
	if err := json.Unmarshal([]byte(msg), &l); err != nil {
		u.infoLog.PrintFunc()(fmt.Sprintf("λ %s", msg))
		return
	}
	if l.Level == levelDebug && !debugLogLevelEnabled {
		return
	}
	c := u.levelColor(l.Level)
	c.PrintlnFunc()(fmt.Sprintf("λ %s", l.Msg))
}

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

func printStack(w io.Writer, err error) {
	if _, ok := err.(stackTracer); !ok {
		return
	}
	inner := err
	stackCounter := 1
	for {
		if st, ok := inner.(stackTracer); ok {
			for i, f := range st.StackTrace() {
				if i == 1 {
					fmt.Fprintf(w, "%d %s\n", stackCounter, inner)
					fmt.Fprintf(w, "\t%+v\n", f)
					stackCounter++
					break
				}
			}
		}
		c, ok := inner.(causer)
		if !ok {
			break
		}
		inner = c.Cause()
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type causer interface {
	Cause() error
}

type UserError struct {
	msg   string
	cause error
}

func NewUserError(err error, msg string) *UserError {
	return &UserError{
		msg:   msg,
		cause: err,
	}
}

func (e *UserError) Unwrap() error {
	return e.cause
}

func (e *UserError) Cause() error {
	return e.cause
}

func (e *UserError) Error() string {
	if e.cause != nil {
		return e.msg + ": " + e.cause.Error()
	}
	return e.msg
}

func (e *UserError) Message() string {
	return e.msg
}

func Wrap(err error, msg ...string) error {
	if len(msg) == 0 {
		return errors.WithStack(err)
	}
	return errors.Wrap(err, msg[0])
}

func WithStack(err error) error {
	return errors.WithStack(err)
}

func WithUserMessage(err error, msg string) error {
	return errors.WithStack(NewUserError(err, msg))
}

func IsUserError(err error) bool {
	var ue *UserError
	return errors.As(err, &ue)
}
