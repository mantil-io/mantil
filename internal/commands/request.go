package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/atoz-technology/mantil-cli/internal/stream"
	"github.com/atoz-technology/mantil.go/pkg/logs"
	"github.com/nats-io/nats.go"
)

func BackendRequest(method string, req interface{}, rsp interface{}) error {
	backendURL, err := BackendURL()
	if err != nil {
		return fmt.Errorf("could not get backend url - %v", err)
	}
	url := fmt.Sprintf("%s/%s", backendURL, method)
	buf, err := json.Marshal(req)
	if err != nil {
		return err
	}
	inbox := nats.NewInbox()
	rspChan := make(chan *http.Response)
	// invoke lambda asynchronously
	go func() {
		httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
		if err != nil {
			log.Println(err)
			rspChan <- nil
			return
		}
		httpReq.Header.Set("x-nats-inbox", inbox)
		rsp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			log.Println(err)
			rspChan <- nil
			return
		}
		rspChan <- rsp
	}()
	// wait for log messages
	stream.Subscribe(inbox, func(nm *nats.Msg) {
		log.Print(string(nm.Data))
	})
	// wait for response
	httpRsp := <-rspChan
	if httpRsp == nil || rsp == nil {
		return nil
	}
	defer httpRsp.Body.Close()
	buf, err = ioutil.ReadAll(httpRsp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(buf, rsp)
	if err != nil {
		return err
	}
	return nil
}

func PrintProjectRequest(url string, req string, includeHeaders, includeLogs bool) error {
	buf := []byte(req)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	if includeLogs {
		wait, err := logListener(httpReq)
		if err != nil {
			return err
		}
		defer wait()
	}
	httpRsp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpRsp.Body.Close()

	fmt.Println(httpRsp.Status)
	if includeHeaders {
		printRspHeaders(httpRsp)
		fmt.Println()
	} else if !isSuccessfulResponse(httpRsp) {
		printApiErrorHeader(httpRsp)
	}

	buf, err = ioutil.ReadAll(httpRsp.Body)
	if err != nil {
		return err
	}
	if string(buf) != "" {
		fmt.Printf("%s\n", string(buf))
	}
	return nil
}

func isSuccessfulResponse(rsp *http.Response) bool {
	return strings.HasPrefix(rsp.Status, "2")
}

func printRspHeaders(rsp *http.Response) {
	for k, v := range rsp.Header {
		fmt.Printf("%s: %s\n", k, strings.Join(v, ","))
	}
}

func printApiErrorHeader(rsp *http.Response) {
	header := "X-Api-Error"
	apiErr := rsp.Header.Get(header)
	if apiErr != "" {
		fmt.Printf("%s: %s\n", header, apiErr)
	}
}

func logListener(req *http.Request) (wait func(), err error) {
	l := logs.NewListener()
	req.Header.Add(logs.InboxHeaderKey, l.Subject())
	wait, err = l.Listen(context.Background(), func(msg string) error {
		fmt.Print(msg)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return wait, nil
}

type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}
