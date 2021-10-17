// Slack reference: https://atoz-technology.slack.com/archives/C024QHF6ZUN/p1633107324326000?thread_ts=1633010861.297400&cid=C024QHF6ZUN
package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
)

var (
	logFile *os.File
	logs    *log.Logger
	errs    *log.Logger
)

func Open() error {
	fn := fmt.Sprintf("/tmp/mantil-%s.log", time.Now().Format("2006-01-02"))
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	logs = log.New(f, "", log.LstdFlags|log.Lmicroseconds|log.Llongfile)
	errs = log.New(f, "[ERROR] ", log.LstdFlags|log.Lmicroseconds|log.Llongfile|log.Lmsgprefix)
	logFile = f
	return nil
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

func PrintfWithCallDepth(calldepth int, format string, v ...interface{}) {
	if logFile == nil {
		return
	}
	logs.Output(calldepth+2, fmt.Sprintf(format, v...))
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
	if err == nil {
		return nil
	}
	if len(msg) == 0 {
		return errors.WithStack(err)
	}
	return errors.Wrap(err, msg[0])
}

// WithUserMessage propagate error with wrapping it in UserError.
// That message will be shown to the Mantil user.
func WithUserMessage(err error, format string, v ...interface{}) error {
	msg := fmt.Sprintf(format, v...)
	return errors.WithStack(newUserError(err, msg))
}

// IsUserError checks whether provided error is of UserError type
func IsUserError(err error) bool {
	var ue *UserError
	return errors.As(err, &ue)
}
