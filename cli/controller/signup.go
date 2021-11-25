package controller

import (
	"fmt"
	"net/mail"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/secret"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/domain/signup"
)

func Register() error {
	rr, err := survey()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return nil
		}
		return log.Wrap(err)
	}
	if err := backend.Signup().Register(rr); err != nil {
		log.Wrap(err)
	}
	ui.Info("Activation token is sent to %s. Please check your email to finalize registration.", rr.Email)
	return nil
}

func Activate(activationCode string) error {
	fs, err := newStore()
	if err != nil {
		return err
	}
	jwt, err := backend.Signup().Activate(signup.NewActivateRequest(activationCode, fs.Workspace().ID))
	if err != nil {
		return log.Wrap(err)
	}
	claims, err := signup.Validate(jwt, secret.SignupPublicKey)
	if err != nil {
		return log.Wrap(err)
	}
	log.SetClaims(claims)
	if err := domain.StoreActivationToken(jwt); err != nil {
		return log.Wrap(err)
	}
	ui.Info("Activation successful. Enjoy building with Mantil!")
	return nil
}

func IsActivated() bool {
	jwt, err := domain.ReadActivationToken()
	if err != nil {
		log.Printf("activation token not found")
		return false
	}
	claims, err := signup.Validate(jwt, secret.SignupPublicKey)
	if err != nil {
		log.Errorf("failed to validate activation token %s", err)
		return false
	}
	log.SetClaims(claims)
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
