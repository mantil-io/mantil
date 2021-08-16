package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

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

func PrintProjectRequest(url string, req string) error {
	buf := []byte(req)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	httpRsp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpRsp.Body.Close()

	fmt.Println(httpRsp.Status)
	if isSuccessfulResponse(httpRsp) {
		buf, err = ioutil.ReadAll(httpRsp.Body)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", string(buf))
	} else {
		apiErr := httpRsp.Header.Get("X-Api-Error")
		if apiErr != "" {
			fmt.Printf("X-Api-Error: %s\n", apiErr)
		}
	}
	return nil
}

func isSuccessfulResponse(rsp *http.Response) bool {
	return strings.HasPrefix(rsp.Status, "2")
}

type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}
