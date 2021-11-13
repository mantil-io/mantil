package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/mail"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/signup"
)

func Register() error {
	rr, err := survey()
	if err != nil {
		return log.Wrap(err)
	}
	if err := signupEndpoint.Call("register", rr, nil); err != nil {
		log.Wrap(err)
	}
	ui.Info("Registration request sent")
	//TODO: you will recive email with ...
	return nil
}

func Activate(id string) error {
	machineID := domain.MachineID()
	ar := signup.ActivateRequest{
		ID:        id,
		MachineID: machineID,
	}

	var jwt string
	if err := signupEndpoint.Call("activate", ar, &jwt); err != nil {
		return log.Wrap(err)
	}
	if err := signup.Validate(jwt, machineID); err != nil {
		return log.Wrap(err, "token not valid")
	}
	if err := domain.StoreActivationToken(jwt); err != nil {
		return log.Wrap(err)
	}
	ui.Info("Activation successful")
	return nil
}

func IsActivated() bool {
	jwt, err := domain.ReadActivationToken()
	if err != nil {
		log.Error(err)
		return false
	}
	if err := signup.Validate(jwt, domain.MachineID()); err != nil {
		log.Error(err)
		return false
	}
	return true
}

func survey() (rr signup.RegisterRequest, err error) {
	prompt := promptui.Prompt{
		Label: "1/4 First things first, what is your name?",
		Validate: func(name string) error {
			if name == "" {
				return fmt.Errorf("name is missing")
			}
			return nil
		},
	}
	rr.Name, err = prompt.Run()
	if err != nil {
		return
	}

	prompt = promptui.Prompt{
		Label: "2/4 And your email address?",
		Validate: func(email string) error {
			_, err = mail.ParseAddress(email)
			if err != nil {
				return fmt.Errorf("email validation failed")
			}
			return nil
		},
	}
	rr.Email, err = prompt.Run()
	if err != nil {
		return
	}

	ps := promptui.Select{
		Label: "3/4 Great! Now what do you do?",
		Items: []string{"Software Engineer", "DevOps Engineer", "Team Lead", "VP of Engineering/CTO", "Other"},
	}
	_, rr.Position, err = ps.Run()
	if err != nil {
		return
	}

	ps = promptui.Select{
		Label: "4/4 Lastly, how big is your development organization?",
		Items: []string{"Only me", "2-10", "11-30", "31-70", "71+"},
	}
	_, rr.OrganizationSize, err = ps.Run()
	return
}

var signupEndpoint = apiEndpoint{url: "https://4fc99dc1lf.execute-api.eu-central-1.amazonaws.com/signup"}

type apiEndpoint struct {
	url string
}

func (a *apiEndpoint) Call(method string, req, rsp interface{}) error {
	buf, _ := json.Marshal(req)
	url := a.url + "/" + method
	httpRsp, err := http.Post(url, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return log.Wrap(err)
	}
	if err != nil {
		return log.Wrap(err)
	}
	defer httpRsp.Body.Close()
	if httpRsp.StatusCode == http.StatusNoContent {
		return nil
	}
	if httpRsp.StatusCode != http.StatusOK {
		if apiErr := httpRsp.Header.Get("X-Api-Error"); apiErr != "" {
			return log.Wrapf(apiErr)
		}
		return log.Wrapf("request failed with status code %d", httpRsp.StatusCode)
	}
	if rsp != nil {
		buf, err := ioutil.ReadAll(httpRsp.Body)
		if err != nil {
			return log.Wrap(err)
		}

		switch v := rsp.(type) {
		case []byte:
			rsp = buf
		case *string:
			*v = string(buf)
		default:
			if err := json.Unmarshal(buf, rsp); err != nil {
				return log.Wrap(err)
			}
		}
	}
	return nil
}
