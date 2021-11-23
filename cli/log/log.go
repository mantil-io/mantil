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
	"github.com/mantil-io/mantil/domain/signup"
	"github.com/pkg/errors"
)

var (
	logFile    *os.File
	logs       *log.Logger
	errs       *log.Logger
	collector  = newEventsCollector()
	cliCommand = &domain.CliCommand{}
)

func newEventsCollector() *eventsCollector {
	ec := eventsCollector{
		connectDone: make(chan struct{}),
		store:       newEventsStore(),
	}

	go func() {
		defer close(ec.connectDone)
		p, err := net.NewPublisher(secret.EventPublisherCreds)
		if err != nil {
			Error(fmt.Errorf("failed to start event publisher %w", err))
			return
		}
		ec.publisher = p
	}()
	return &ec
}

type eventsCollector struct {
	publisher   *net.Publisher
	connectDone chan struct{}
	store       *eventsStore
}

func (c *eventsCollector) send() error {
	cliCommand.End()
	buf, err := cliCommand.Marshal()
	if err != nil {
		c.store.push(buf)
		return Wrap(err)
	}
	// wait for net connection to finish
	<-c.connectDone
	if c.publisher == nil {
		c.store.push(buf)
		return fmt.Errorf("publisher not found")
	}
	if err := c.publisher.Pub(buf); err != nil {
		c.store.push(buf)
		return Wrap(err)
	}
	return nil
}

func (c *eventsCollector) close() error {
	if err := c.send(); err != nil {
		return c.store.store()
	}
	if c.publisher == nil {
		return c.store.store()
	}

	// clear already sent events
	if err := c.store.clear(); err != nil {
		return err
	}
	// find events on disk and send them also
	if err := c.store.restore(); err != nil {
		return err
	}
	var err error
	for _, v := range c.store.events {
		err = c.publisher.Pub(v)
	}
	if err == nil {
		if err := c.store.clear(); err != nil {
			return err
		}
	}
	return c.publisher.Close()
}

func Open() error {
	logsDir, err := LogsDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(logsDir, os.ModePerm); err != nil {
		return Wrap(fmt.Errorf("failed to create application logs dir %s, error %w", logsDir, err))
	}
	logFilePath := filepath.Join(logsDir, LogFileForDate(time.Now()))

	f, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	logs = log.New(f, "", log.LstdFlags|log.Lmicroseconds|log.Llongfile)
	errs = log.New(f, "[ERROR] ", log.LstdFlags|log.Lmicroseconds|log.Llongfile|log.Lmsgprefix)
	logFile = f

	cliCommand.Start()
	logWorkspace()
	collector = newEventsCollector()
	return nil
}

func Close() {
	if err := collector.close(); err != nil {
		Error(err)
	}
	if logFile != nil {
		fmt.Fprintf(logFile, "\n")
		logFile.Close()
	}
}

func LogsDir() (string, error) {
	appConfigDir, err := domain.AppConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(appConfigDir, "logs"), nil
}

func LogFileForDate(date time.Time) string {
	return fmt.Sprintf("%s.log", date.Format("2006-01-02"))
}

func SetStage(fs *domain.FileStore, p *domain.Project, s *domain.Stage) {
	cliCommand.Workspace = fs.AsCliWorkspace()
	cliCommand.Project = p.AsCliProject()
	cliCommand.Stage = s.AsCliStage()
}

func SetClaims(claims *signup.TokenClaims) {
	if claims == nil {
		return
	}
	cliCommand.User = &domain.CliUser{
		ID:             claims.ActivationCode,
		ActivationCode: claims.ActivationCode,
		Email:          claims.Email,
	}
}

func Event(e domain.Event) {
	cliCommand.Add(e)
}

func SendEvents() error {
	err := collector.send()
	cliCommand.Clear()
	logWorkspace()
	return err
}

func logWorkspace() {
	fs, err := domain.NewSingleDeveloperWorkspaceStore()
	if err != nil {
		return
	}
	project := fs.Project()
	if project == nil {
		SetStage(fs, nil, nil)
	}
	SetStage(fs, project, fs.DefaultStage())
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

// ForEachInnerError executes callback for each inner message of err
// First occured, the deapest in the stack is called first.
func ForEachInnerError(err error, cb func(error)) {
	if err == nil {
		return
	}
	// collect errors list
	var errList []error
	for {
		if ue, ok := err.(*innerError); ok {
			errList = append(errList, ue)
		}
		c, ok := err.(causer)
		if !ok {
			break
		}
		err = c.Cause()
	}
	if err != nil {
		if _, ok := err.(*innerError); !ok {
			errList = append(errList, err)
		}
	}
	// call callback in reverse order
	for i := len(errList) - 1; i >= 0; i-- {
		cb(errList[i])
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type causer interface {
	Cause() error
}

type innerError struct {
	msg   string
	cause error
}

func newInnerError(err error, msg string) *innerError {
	return &innerError{
		msg:   msg,
		cause: err,
	}
}

func (e *innerError) Unwrap() error {
	return e.cause
}

func (e *innerError) Cause() error {
	return e.cause
}

func (e *innerError) Error() string {
	return e.msg
}

func (e *innerError) Message() string {
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
	return errors.WithStack(newInnerError(err, msg))
}

func Wrapf(format string, v ...interface{}) error {
	msg := fmt.Sprintf(format, v...)
	return errors.WithStack(newInnerError(nil, msg))
}

type GoBuildError struct {
	Name  string
	Dir   string
	Lines []string
}

func (e *GoBuildError) Error() string {
	return strings.Join(e.Lines, "\n")
}

var NotActivatedError = fmt.Errorf("not activated")
