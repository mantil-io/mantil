// Slack reference: https://atoz-technology.slack.com/archives/C024QHF6ZUN/p1633107324326000?thread_ts=1633010861.297400&cid=C024QHF6ZUN
package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/mantil-io/mantil/cli/log/net"
	"github.com/mantil-io/mantil/cli/secret"
	"github.com/mantil-io/mantil/domain"
	"github.com/pkg/errors"
)

var (
	logFile              *os.File
	logs                 *log.Logger
	errs                 *log.Logger
	cliCommand           *domain.CliCommand
	publisher            *net.Publisher
	publisherConnectDone chan struct{}
)

func Open() error {
	appConfigDir, err := domain.AppConfigDir()
	if err != nil {
		return err
	}

	logsDir := filepath.Join(appConfigDir, "logs")
	if err := os.MkdirAll(logsDir, os.ModePerm); err != nil {
		return Wrap(fmt.Errorf("failed to create application logs dir %s, error %w", logsDir, err))
	}
	logFilePath := filepath.Join(logsDir, time.Now().Format("2006-01-02")+".log")

	f, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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
	publisherConnectDone := make(chan struct{})
	var cc domain.CliCommand
	cc.Start()

	// start net connection in another thread
	// it will hopefully be ready by end of the commnad
	// trying to avoid small wait time to establish connection at the end of every command
	go func() {
		defer close(publisherConnectDone)
		p, err := net.NewPublisher(secret.EventPublisherCreds)
		if err != nil {
			Error(err)
			return
		}
		publisher = p
	}()
	cliCommand = &cc
}

func Close() {
	if err := sendEvents(); err != nil {
		Error(err)
	}
	if publisher != nil {
		if err := publisher.Close(); err != nil {
			Error(err)
		}
	}
	if logFile != nil {
		fmt.Fprintf(logFile, "\n")
		logFile.Close()
	}
}

func SetStage(w *domain.Workspace, p *domain.Project, s *domain.Stage) {
	if w != nil {
		cliCommand.Workspace.Name = w.Name
		cliCommand.Workspace.Nodes = len(w.Nodes)
	}
	if p != nil {
		cliCommand.Project.Name = p.Name
		cliCommand.Project.Stages = p.NumberOfStages()
		cliCommand.Project.Nodes = p.NumberOfNodes()
		cliCommand.Project.AWSAccounts = p.NumberOfAWSAccounts()
	}
	if s != nil {
		cliCommand.Stage.Name = s.Name
		cliCommand.Stage.Node = s.NodeName
		cliCommand.Stage.Functions = len(s.Functions)
		if s.Public != nil {
			cliCommand.Stage.PublicFolders = len(s.Public.Sites)
		}
	}
}

func sendEvents() error {
	cliCommand.End()
	buf, err := cliCommand.Marshal()
	if err != nil {
		return Wrap(err)
	}
	// wait for net connection to finish
	<-publisherConnectDone
	if publisher == nil {
		return fmt.Errorf("publisher not found")
	}
	if err := publisher.Pub(buf); err != nil {
		return Wrap(err)
	}
	return nil
}

func Event(e domain.Event) {
	cliCommand.Add(e)
}

func SendEvents() error {
	err := sendEvents()
	cliCommand.Clear()
	return err
}

func Printf(format string, v ...interface{}) {
	if logFile == nil {
		return
	}
	logs.Output(2, fmt.Sprintf(format, v...))
}

func Signal(name string) {
	if logFile == nil {
		return
	}
	bb := bytes.NewBuffer(nil)
	Printf("signal %s", name)
	pprof.Lookup("goroutine").WriteTo(bb, 1)
	buf := bb.Bytes()
	logFile.Write(buf)
	cliCommand.Add(domain.Event{Signal: &domain.Signal{Name: name, Stack: string(buf)}})
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
		if inner == nil {
			break
		}
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
