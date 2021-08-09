package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/atoz-technology/mantil-cli/internal/stream"
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

type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}
