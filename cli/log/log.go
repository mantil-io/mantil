// Slack reference: https://atoz-technology.slack.com/archives/C024QHF6ZUN/p1633107324326000?thread_ts=1633010861.297400&cid=C024QHF6ZUN
package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mantil-io/mantil/cli/log/net"
	"github.com/mantil-io/mantil/cli/secret"
	"github.com/mantil-io/mantil/domain"
	"github.com/pkg/errors"
)

var (
	logFile        *os.File
	logs           *log.Logger
	errs           *log.Logger
	cliCommand     *domain.CliCommand
	eventPublisher chan func([]byte) error
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
	startEventCollector()
	return nil
}

func startEventCollector() {
	var cc domain.CliCommand
	cc.Start()
	cc.Args = os.Args
	cc.Version = domain.Version()
	// TODO add other attributes

	// start net connection in another thread
	// it will hopefully be ready by end of the commnad
	// trying to avoid small wait time to establish connection at the end of every command
	eventPublisher = make(chan func([]byte) error, 1)
	go func() {
		p, err := net.NewPublisher(secret.EventPublisherCreds)
		defer close(eventPublisher)
		if err != nil {
			Error(err)
			return
		}
		eventPublisher <- p.Pub
	}()
	cliCommand = &cc
}

func Close() {
	if err := sendEvents(); err != nil {
		Error(err)
	}
	if logFile != nil {
		fmt.Fprintf(logFile, "\n")
		logFile.Close()
	}
}

func sendEvents() error {
	cliCommand.End()
	buf, err := cliCommand.Marshal()
	if err != nil {
		return Wrap(err)
	}
	// wait for net connection to finish
	ep := <-eventPublisher
	if ep == nil {
		return fmt.Errorf("publisher not found")
	}
	if err := ep(buf); err != nil {
		return Wrap(err)
	}
	return nil
}

func Event(e domain.Event) {
	cliCommand.Add(e)
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
		cliCommand.AddError(domain.CliError{
			Error: err.Error(),
			Type:  fmt.Sprintf("%T", err),
		})
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

					cliCommand.AddError(domain.CliError{
						//Type:         fmt.Sprintf("%T", inner), // it is always *errors.withStack
						Error:        inner.Error(),
						SourceFile:   fmt.Sprintf("%v", f),
						FunctionName: fmt.Sprintf("%n", f),
					})

					stackCounter++
					break
				}
			}
		}
		c, ok := inner.(causer)
		if !ok {
			cliCommand.AddError(domain.CliError{
				Error: inner.Error(),
				Type:  fmt.Sprintf("%T", inner),
			})
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
func Wrap(err error, args ...interface{}) error {
	if err == nil {
		return nil
	}
	if len(args) == 0 {
		return errors.WithStack(err)
	}
	msg, ok := args[0].(string)
	if !ok {
		return errors.WithStack(err)
	}
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args[1:]...)
	}
	return errors.WithStack(newUserError(err, msg))
	//return errors.Wrap(err, msg)
}

func Wrapf(format string, v ...interface{}) error {
	msg := fmt.Sprintf(format, v...)
	return errors.WithStack(newUserError(nil, msg))
}

// // WithUserMessage propagate error with wrapping it in UserError.
// // That message will be shown to the Mantil user.
// func WithUserMessage(err error, format string, v ...interface{}) error {
// 	if err == nil {
// 		return nil
// 	}
// 	msg := fmt.Sprintf(format, v...)
// 	return errors.WithStack(newUserError(err, msg))
// }

// IsUserError checks whether provided error is of UserError type
func IsUserError(err error) bool {
	var ue *UserError
	return errors.As(err, &ue)
}

type GoBuildError struct {
	Name  string
	Dir   string
	Lines []string
}

func (e *GoBuildError) Error() string {
	return strings.Join(e.Lines, "\n")
}
