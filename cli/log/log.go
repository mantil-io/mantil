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

func (u *UILogger) DisableColor() {
	print := func(a ...interface{}) {
		fmt.Println(a...)
	}
	u.infoLog = print
	u.debugLog = print
	u.noticeLog = print
	u.errorLog = func(a ...interface{}) {
		fmt.Printf("Error: %s\n", fmt.Sprint(a...))
	}
	u.fatalLog = func(a ...interface{}) {
		fmt.Printf("Fatal: %s\n", fmt.Sprint(a...))
	}
}

var (
	logFile *os.File
	logs    *log.Logger
	errs    *log.Logger
)

func init() {
	UI = &UILogger{
		infoLog:   color.New().PrintlnFunc(),
		debugLog:  color.New(color.Faint).PrintlnFunc(),
		noticeLog: color.New(color.FgGreen, color.Bold).PrintlnFunc(),
		errorLog:  color.New(color.FgRed).PrintlnFunc(),
		fatalLog:  color.New(color.FgRed, color.Bold).PrintlnFunc(),
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
					// zero stack entry is from this package
					// from Wrap or WithUserMessage method
					// so the real caller where error is wrapped is as stack index 1
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

func newUserError(err error, msg string) *UserError {
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

// Wrap each error with the stack (file and line) where the error is wrapped
func Wrap(err error, msg ...string) error {
	if len(msg) == 0 {
		return errors.WithStack(err)
	}
	return errors.Wrap(err, msg[0])
}

// WithUserMessage propagate error with wrapping it in UserError.
// That message will be shown to the Mantil user.
func WithUserMessage(err error, msg string) error {
	return errors.WithStack(newUserError(err, msg))
}

// IsUserError checks whether provided error is of UserError type
func IsUserError(err error) bool {
	var ue *UserError
	return errors.As(err, &ue)
}
