package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mantil-io/mantil.go/pkg/streaming/nats"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
)

type Backend struct {
	endpoint    string
	authToken   string
	includeLogs bool
	logSink     func(chan []byte)
	onRsp       func(*http.Response) error
}

func New(endpoint, authToken string) *Backend {
	return &Backend{
		endpoint:    endpoint,
		includeLogs: true,
		authToken:   authToken,
		logSink:     backendLogsSink,
	}
}

func Project(endpoint string, includeHeaders, includeLogs bool) *Backend {
	return &Backend{
		endpoint:    endpoint,
		includeLogs: includeLogs,
		logSink:     projectLogsSink,
		onRsp:       buildShowResponseHandler(includeHeaders),
	}
}

func (b *Backend) Call(method string, req interface{}, rsp interface{}) error {
	httpReq, err := b.newHTTPRequest(method, req)
	if err != nil {
		return log.Wrap(err)
	}

	var listener *listener
	if b.includeLogs {
		var err error
		listener, err = newListener(httpReq, rsp, b.logSink)
		if err != nil {
			log.Errorf("failed to start log listener - %v", err)
			// fallback to getting response from http
		}
	}

	httpRsp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return log.Wrap(err, fmt.Sprintf("failed to make http request to %s", httpReq.URL))
	}
	defer httpRsp.Body.Close()

	if b.onRsp != nil {
		if listener != nil {
			defer listener.responseStatus()
		}
		return b.onRsp(httpRsp)
	}

	if listener != nil {
		remoteErr, localErr := listener.responseStatus() // wait for response to arrive
		if localErr == nil {
			return remoteErr
		}
		// log error and fallback to http response
		log.Errorf("logs callback error - %v", localErr)
	}

	if err := checkResponse(httpRsp); err != nil {
		return log.Wrap(err)
	}
	return unmarshalBody(httpRsp, rsp)
}

func checkResponse(httpRsp *http.Response) error {
	if apiErr := httpRsp.Header.Get("X-Api-Error"); apiErr != "" {
		log.Errorf("api error %s", apiErr)
		return log.Wrap(fmt.Errorf(apiErr))
	}
	if !(httpRsp.StatusCode == http.StatusOK ||
		httpRsp.StatusCode == http.StatusNoContent) {
		return log.Wrap(fmt.Errorf("request http status %d", httpRsp.StatusCode))
	}
	return nil
}

func (b *Backend) url(method string) string {
	return fmt.Sprintf("%s/%s", b.endpoint, method)
}

func (b *Backend) newHTTPRequest(method string, req interface{}) (*http.Request, error) {
	buf, err := b.marshal(req)
	if err != nil {
		return nil, log.Wrap(err, "failed to marshal request object")
	}
	httpReq, err := http.NewRequest("POST", b.url(method), bytes.NewBuffer(buf))
	if err != nil {
		return nil, log.Wrap(err, "could not create request")
	}
	httpReq.Header.Add(domain.AccessTokenHeader, b.authToken)
	return httpReq, nil
}

func (b *Backend) marshal(o interface{}) ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	switch v := o.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(o)
	}
}

func unmarshalBody(httpRsp *http.Response, rsp interface{}) error {
	if rsp == nil {
		return nil
	}
	httpBody, err := ioutil.ReadAll(httpRsp.Body)
	if err != nil {
		return log.Wrap(err, "failed to read http response body")
	}
	err = json.Unmarshal(httpBody, rsp)
	if err != nil {
		return log.Wrap(err, "failed to unmarshal http response")
	}
	return nil
}

type listener struct {
	natsListener *nats.LambdaListener
	errc         chan error
	logSink      func(chan []byte)
}

func newListener(httpReq *http.Request, rsp interface{}, logSink func(chan []byte)) (*listener, error) {
	nl, err := nats.NewLambdaListener()
	if err != nil {
		return nil, err
	}
	l := &listener{
		natsListener: nl,
		errc:         make(chan error, 1),
		logSink:      logSink,
	}
	if err := l.startLogsLoop(); err != nil {
		return nil, err
	}
	go l.waitForResponse(rsp)
	l.setHTTPHeaders(httpReq)
	return l, nil
}

// remote error, local error
func (l *listener) responseStatus() (error, error) {
	err := <-l.errc // wait to get reponse

	if err == nil {
		return nil, nil // callback succeeded rsp is filled
	}
	var rerr *nats.ErrRemoteError
	if errors.As(err, &rerr) {
		return err, nil // callback ok remote retured error
	}
	log.Errorf("logs callback error - %v", err)
	return nil, err
}

func (l *listener) setHTTPHeaders(httpReq *http.Request) {
	for k, v := range l.natsListener.Headers() {
		httpReq.Header.Add(k, v)
	}
}

func (l *listener) waitForResponse(rsp interface{}) {
	switch v := rsp.(type) {
	case *bytes.Buffer:
		buf, err := l.natsListener.RawResponse(context.Background())
		v.Write(buf)
		l.errc <- err
	default:
		l.errc <- l.natsListener.Response(context.Background(), rsp)
	}
	l.errc <- l.natsListener.Response(context.Background(), rsp)
	close(l.errc)
}

func (l *listener) startLogsLoop() error {
	ctx := context.Background()
	logsCh, err := l.natsListener.Logs(ctx)
	if err != nil {
		return err
	}
	go l.logSink(logsCh)
	return nil
}

func backendLogsSink(logsCh chan []byte) {
	tp := ui.NewTerraformProgress()
	for buf := range logsCh {
		msg := string(buf)
		if ok := tp.Parse(msg); ok {
			continue
		}
		if strings.HasPrefix(msg, "EVENT: ") {
			ui.Info(strings.TrimPrefix(msg, "EVENT: "))
		}
		log.Printf(msg)
	}
}

func projectLogsSink(logsCh chan []byte) {
	for buf := range logsCh {
		ui.Info("λ %s", buf)
	}
}

func buildShowResponseHandler(includeHeaders bool) func(httpRsp *http.Response) error {
	return func(httpRsp *http.Response) error {
		if isSuccessfulResponse(httpRsp) {
			ui.Notice(httpRsp.Status)
		} else {
			ui.Errorf(httpRsp.Status)
		}

		if includeHeaders {
			printRspHeaders(httpRsp)
			ui.Info("")
		} else if !isSuccessfulResponse(httpRsp) {
			printApiErrorHeader(httpRsp)
		}

		buf, err := ioutil.ReadAll(httpRsp.Body)
		if err != nil {
			return err
		}
		if string(buf) != "" {
			dst := &bytes.Buffer{}
			if err := json.Indent(dst, buf, "", "   "); err != nil {
				ui.Info(string(buf))
			} else {
				ui.Info(dst.String())
			}
		}
		return nil
	}
}

func isSuccessfulResponse(rsp *http.Response) bool {
	return strings.HasPrefix(rsp.Status, "2")
}

func printRspHeaders(rsp *http.Response) {
	for k, v := range rsp.Header {
		ui.Info("%s: %s", k, strings.Join(v, ","))
	}
}

func printApiErrorHeader(rsp *http.Response) {
	header := "X-Api-Error"
	apiErr := rsp.Header.Get(header)
	if apiErr != "" {
		ui.Info("%s: %s", header, apiErr)
	}
}

type LambdaCaller struct {
	invoker      Invoker
	functionName string
}

type Invoker interface {
	Invoke(name string, req, rsp interface{}, headers map[string]string) error
}

func Lambda(invoker Invoker, functionName string) *LambdaCaller {
	return &LambdaCaller{
		invoker:      invoker,
		functionName: functionName,
	}
}

func (l *LambdaCaller) Call(method string, req, rsp interface{}) error {
	lsn, err := newLambdaListener(rsp)

	var payload []byte
	if req != nil {
		var err error
		payload, err = json.Marshal(req)
		if err != nil {
			return err
		}
	}

	reqWithURI := struct {
		URI     string
		Payload []byte
	}{
		URI:     method,
		Payload: payload,
	}
	err = l.invoker.Invoke(l.functionName, reqWithURI, rsp, lsn.natsListener.Headers())
	if err != nil {
		return err
	}
	remoteErr, localErr := lsn.responseStatus()
	if localErr == nil {
		return remoteErr
	}
	// log error and fallback to http response
	log.Errorf("logs callback error - %v", localErr)
	return nil
}

func newLambdaListener(rsp interface{}) (*listener, error) {
	nl, err := nats.NewLambdaListener()
	if err != nil {
		return nil, err
	}
	l := &listener{
		natsListener: nl,
		errc:         make(chan error, 1),
		logSink:      backendLogsSink,
	}
	if err := l.startLogsLoop(); err != nil {
		return nil, err
	}
	go l.waitForResponse(rsp)
	return l, nil
}
