package signup

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil/backend/secret"
	"github.com/mantil-io/mantil/domain/signup"
	"github.com/mantil-io/mantil/texts"
)

const (
	registrationsPartition = "registrations"
	activationsPartition   = "activations"
	workspacesPartition    = "workspaces"
)

var (
	internalServerError = fmt.Errorf("internal server error")
	badRequestError     = fmt.Errorf("bad request")
)

type Signup struct {
	kv      *mantil.KV
	noEmail bool
}

func New() *Signup {
	return &Signup{}
}

func (r *Signup) connectKV() error {
	if r.kv != nil {
		return nil
	}

	kv, err := mantil.NewKV(registrationsPartition)
	if err != nil {
		log.Printf("mantil.NewKV failed: %s", err)
		return internalServerError
	}
	r.kv = kv
	return nil
}

func (r *Signup) put(rec signup.Record) error {
	if err := r.connectKV(); err != nil {
		return internalServerError
	}

	if err := r.kv.Put(rec.ID, rec); err != nil {
		log.Printf("kv.Put failed: %s", err)
		return internalServerError
	}
	return nil
}

func (r *Signup) get(id string) (signup.Record, error) {
	var rec signup.Record

	if err := r.connectKV(); err != nil {
		return rec, internalServerError
	}

	if err := r.kv.Get(id, &rec); err != nil {
		log.Printf("kv.Get failed: %s", err)
		return rec, fmt.Errorf("activation token not found")
	}

	return rec, nil
}

func (r *Signup) register(ctx context.Context, req signup.RegisterRequest) (*signup.Record, error) {
	if !req.Valid() {
		return nil, badRequestError
	}
	rec := req.AsRecord()
	rec.RemoteIP = remoteIP(ctx)
	if err := r.put(rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *Signup) Register(ctx context.Context, req signup.RegisterRequest) error {
	rec, err := r.register(ctx, req)
	if err != nil {
		return err
	}
	if err := r.sendActivationCode(rec.Email, rec.Name, rec.ActivationCode); err != nil {
		return internalServerError
	}
	return nil
}

func (r *Signup) activate(ctx context.Context, req signup.ActivateRequest) (*signup.Record, error) {
	if !req.Valid() {
		return nil, badRequestError
	}
	rec, err := r.get(req.Code())
	if err != nil {
		return nil, err
	}

	if rec.Activated() {
		if rec.ActivatedFor(req.MachineID) {
			return &rec, nil
		}
		return nil, fmt.Errorf("token already used on another machine")
	}

	rec.Activate(req)
	token, err := secret.Encode(rec.AsTokenClaims())
	if err != nil {
		log.Printf("failed to encode user token error: %s", err)
		return nil, internalServerError
	}
	rec.Token = token
	rec.RemoteIP = remoteIP(ctx)

	if err := r.put(rec); err != nil {
		return nil, internalServerError
	}
	return &rec, nil
}

func (r *Signup) Activate(ctx context.Context, req signup.ActivateRequest) (string, error) {
	rec, err := r.activate(ctx, req)
	if err != nil {
		return "", err
	}
	if err := r.sendWelcomeMail(rec.Email, rec.Name); err != nil {
		log.Printf("failed to sedn welcome mail error %s", err)
		// do nothing, not critical
	}
	return rec.Token, nil
}

func remoteIP(ctx context.Context) string {
	rc, ok := mantil.FromContext(ctx)
	if !ok {
		return ""
	}
	return rc.Request.RemoteIP()
}

func rawRequest(ctx context.Context) []byte {
	rc, ok := mantil.FromContext(ctx)
	if !ok {
		return nil
	}
	return rc.Request.Raw
}

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
	if r.noEmail {
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

func (r *Signup) typeform(ctx context.Context, req signup.TypeformWebhook) (*signup.Record, error) {
	if !req.Valid() {
		return nil, badRequestError
	}

	rec := req.AsRecord()
	rec.Raw = rawRequest(ctx)
	rec.RemoteIP = remoteIP(ctx)
	if err := r.put(rec); err != nil {
		return nil, err
	}

	return &rec, nil
}

func (r *Signup) Typeform(ctx context.Context, req signup.TypeformWebhook) error {
	rec, err := r.typeform(ctx, req)
	if err != nil {
		return err
	}
	if err := r.sendActivationCode(rec.Email, rec.Name, rec.ActivationCode); err != nil {
		return internalServerError
	}
	return nil
}
