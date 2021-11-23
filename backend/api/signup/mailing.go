package signup

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/mantil-io/mantil/domain/signup"
	"github.com/mantil-io/mantil/texts"
)

func (r *Signup) sendActivationCode(email, name, activationCode string) error {
	toEmail := email
	fromEmail := texts.MailFrom
	subject := texts.ActivationMailSubject
	body, err := texts.ActivationMailBody(name, activationCode)
	if err != nil {
		log.Printf("failed to get mail body: %s", err)
		return err
	}
	return r.sendEmail(fromEmail, toEmail, subject, body)
}

func (r Signup) sendWelcomeMail(email, name string) error {
	toEmail := email
	fromEmail := texts.MailFrom
	subject := texts.WelcomeMailSubject
	body, err := texts.WelcomeMailBody(name)
	if err != nil {
		log.Printf("failed to get mail body: %s", err)
		return err
	}
	return r.sendEmail(fromEmail, toEmail, subject, body)
}

func (r *Signup) sendEmail(fromEmail, toEmail, subject, body string) error {
	if toEmail == signup.TestEmail { // don't send email for integration test
		return nil
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("failed to load configuration: %s", err)
		return err
	}
	cli := ses.NewFromConfig(cfg)

	smi := &ses.SendEmailInput{
		Message: &types.Message{
			Body: &types.Body{
				Text: &types.Content{
					Data: aws.String(body),
				},
			},
			Subject: &types.Content{
				Data: aws.String(subject),
			},
		},
		Destination: &types.Destination{
			ToAddresses: []string{toEmail},
		},
		Source: aws.String(fromEmail),
	}

	if _, err := cli.SendEmail(context.Background(), smi); err != nil {
		log.Printf("send email error: %s", err)
		return err
	}
	return nil
}
