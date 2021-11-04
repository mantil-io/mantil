package backend

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mantil-io/mantil.go/logs"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/secret"
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

func Project(endpoint string, includeLogs bool, cb func(*http.Response) error) *Backend {
	return &Backend{
		endpoint:    endpoint,
		includeLogs: includeLogs,
		logSink:     projectLogsSink,
		onRsp:       cb,
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
		if listener != nil && (httpRsp.StatusCode == http.StatusOK ||
			httpRsp.StatusCode == http.StatusNoContent) {
			_, _ = listener.responseStatus()
		}
		return b.onRsp(httpRsp)
	}

	// if not timeout
	if httpRsp.StatusCode != http.StatusGatewayTimeout &&
		httpRsp.StatusCode != http.StatusServiceUnavailable {
		if err := checkResponse(httpRsp); err != nil {
			return log.Wrap(err)
		}
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
	natsListener *logs.LambdaListener
}

func newListener(httpReq *http.Request, rsp interface{}, logSink func(chan []byte)) (*listener, error) {
	nl, err := logs.NewLambdaListener(logs.ListenerConfig{
		PublisherJWT: secret.LogsPublisherCreds,
		ListenerJWT:  secret.LogsListenerCreds,
		LogSink:      logSink,
		Rsp:          rsp})
	if err != nil {
		return nil, err
	}
	l := &listener{
		natsListener: nl,
	}
	if httpReq != nil {
		l.setHTTPHeaders(httpReq)
	}
	return l, nil
}

// remote error, local error
func (l *listener) responseStatus() (error, error) {
	err := l.natsListener.Done()
	if err == nil {
		return nil, nil // callback succeeded rsp is filled
	}
	var rerr *logs.ErrRemoteError
	if errors.As(err, &rerr) {
		return rerr, nil // callback ok remote retured error
	}
	log.Errorf("logs callback error - %v", err)
	return nil, err
}

func (l *listener) setHTTPHeaders(httpReq *http.Request) {
	for k, v := range l.natsListener.Headers() {
		httpReq.Header.Add(k, v)
	}
}

func backendLogsSink(logsCh chan []byte) {
	tp := ui.NewTerraformProgress()
	for buf := range logsCh {
		msg := string(buf)
		tp.Parse(msg)
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

type LogSinkCallback func(chan []byte)

type LambdaCaller struct {
	invoker      Invoker
	functionName string
	logSink      LogSinkCallback
}

type Invoker interface {
	Invoke(name string, req, rsp interface{}, headers map[string]string) error
}

func Lambda(invoker Invoker, functionName string, logSink LogSinkCallback) *LambdaCaller {
	if logSink == nil {
		logSink = backendLogsSink
	}
	return &LambdaCaller{
		invoker:      invoker,
		functionName: functionName,
		logSink:      logSink,
	}
}

func (l *LambdaCaller) Call(method string, req, rsp interface{}) error {
	lsn, err := newListener(nil, rsp, l.logSink)

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
