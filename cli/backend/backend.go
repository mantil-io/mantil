package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mantil-io/mantil.go/pkg/streaming/nats"
	"github.com/mantil-io/mantil/auth"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/terraform"
)

type Backend struct {
	endpoint  string
	authToken string
}

func New(endpoint, authToken string) *Backend {
	return &Backend{
		endpoint:  endpoint,
		authToken: authToken,
	}
}

func (b *Backend) Call(method string, req interface{}, rsp interface{}) error {
	httpReq, err := b.newHTTPRequest(method, req)
	if err != nil {
		return log.Wrap(err)
	}

	listener, err := newListener(httpReq, rsp)
	if err != nil {
		log.Errorf("failed to start log listener - %v", err)
		// fallback to getting response from http
	}

	httpRsp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return log.Wrap(err, fmt.Sprintf("failed to make http request to %s", httpReq.URL))
	}
	defer httpRsp.Body.Close()

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
	buf, err := json.Marshal(req)
	if err != nil {
		return nil, log.Wrap(err, "failed to marshal request object")
	}
	httpReq, err := http.NewRequest("POST", b.url(method), bytes.NewBuffer(buf))
	if err != nil {
		return nil, log.Wrap(err, "could not create request")
	}
	httpReq.Header.Add(auth.AccessTokenHeader, b.authToken)
	return httpReq, nil
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
}

func newListener(httpReq *http.Request, rsp interface{}) (*listener, error) {
	nl, err := nats.NewLambdaListener()
	if err != nil {
		return nil, err
	}
	l := &listener{
		natsListener: nl,
		errc:         make(chan error, 1),
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
	l.errc <- l.natsListener.Response(context.Background(), rsp)
	close(l.errc)
}

func (l *listener) startLogsLoop() error {
	ctx := context.Background()
	logsCh, err := l.natsListener.Logs(ctx)
	if err != nil {
		return err
	}
	go func() {
		tp := terraform.NewLogParser()
		for buf := range logsCh {
			msg := string(buf)
			if l, ok := tp.Parse(msg); ok {
				if l != "" {
					ui.Info(l)
				}
				log.Printf(msg)
				continue
			}
			ui.Backend(msg)
		}
	}()
	return nil
}
