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
	"github.com/mantil-io/mantil/signup"
	"github.com/mantil-io/mantil/signup/secret"
	"github.com/mantil-io/mantil/texts"
)

const registrationsPartition = "registrations"

var (
	internalServerError = fmt.Errorf("internal server error")
	badRequestError     = fmt.Errorf("bad request")
)

type Signup struct {
	kv *mantil.KV
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

func (r *Signup) Register(ctx context.Context, req signup.RegisterRequest) error {
	if !req.Valid() {
		return badRequestError
	}

	rec := req.AsRecord()
	rec.RemoteIP = remoteIP(ctx)
	if err := r.put(rec); err != nil {
		return err
	}

	if req.Email == signup.TestEmail { // don't send email for integration test
		return nil
	}

	if err := r.sendActivationToken(rec.Email, rec.Name, rec.ID); err != nil {
		return internalServerError
	}

	return nil
}

func (r *Signup) Activate(ctx context.Context, req signup.ActivateRequest) (string, error) {
	if !req.Valid() {
		return "", badRequestError
	}
	rec, err := r.get(req.ID)
	if err != nil {
		return "", err
	}

	if rec.Activated() {
		if rec.ActivatedFor(req.MachineID) {
			return rec.Token, nil
		}
		return "", fmt.Errorf("token already used on another machine")
	}

	rec.Activate(req)
	token, err := secret.Encode(rec.AsTokenClaims())
	if err != nil {
		log.Printf("failed to encode user token error: %s", err)
		return "", internalServerError
	}
	rec.Token = token
	rec.RemoteIP = remoteIP(ctx)

	if err := r.put(rec); err != nil {
		return "", internalServerError
	}

	return token, nil
}

func remoteIP(ctx context.Context) string {
	rc, ok := mantil.FromContext(ctx)
	if !ok {
		return ""
	}
	return rc.Request.RemoteIP()
}

func (r *Signup) sendActivationToken(email, name, id string) error {
	toEmail := email
	fromEmail := texts.ActivationMailFrom
	subject := texts.ActivationMailSubject
	body, err := texts.ActivationMailBody(name, id)
	if err != nil {
		log.Printf("failed to get mail body: %s", err)
		return err
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

func (r *Signup) Typeform(ctx context.Context, req signup.TypeformWebhook) error {
	if !req.Valid() {
		return badRequestError
	}

	rec := req.AsRecord()
	rec.RemoteIP = remoteIP(ctx)
	if err := r.put(rec); err != nil {
		return err
	}

	if err := r.sendActivationToken(rec.Email, rec.Name, rec.ID); err != nil {
		return internalServerError
	}

	return nil
}
