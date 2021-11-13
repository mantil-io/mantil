package register

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil/registration"
	"github.com/mantil-io/mantil/registration/secret"
)

const registrationsPartition = "registrations"

var (
	internalServerError = fmt.Errorf("internal server error")
	badRequestError     = fmt.Errorf("bad request")
)

type Register struct {
	kv *mantil.KV
}

func New() *Register {
	return &Register{}
}

func (r *Register) connectKV() error {
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

func (r *Register) put(rec registration.Record) error {
	if err := r.connectKV(); err != nil {
		return internalServerError
	}

	if err := r.kv.Put(rec.ID, rec); err != nil {
		log.Printf("kv.Put failed: %s", err)
		return internalServerError
	}
	return nil
}

func (r *Register) get(id string) (registration.Record, error) {
	var rec registration.Record

	if err := r.connectKV(); err != nil {
		return rec, internalServerError
	}

	if err := r.kv.Get(id, &rec); err != nil {
		log.Printf("kv.Get failed: %s", err)
		return rec, fmt.Errorf("not found")
	}

	return rec, nil
}

func (r *Register) Register(ctx context.Context, req registration.RegisterRequest) error {
	if !req.Valid() {
		return badRequestError
	}

	rec := req.AsRecord()
	if err := r.put(rec); err != nil {
		return err
	}

	if err := r.notifyByEmail(rec.Email, rec.ID); err != nil {
		return internalServerError
	}

	return nil
}

func (r *Register) Activate(ctx context.Context, req registration.VerifyRequest) (string, error) {
	if !req.Valid() {
		return "", badRequestError
	}
	rec, err := r.get(req.ID)
	if err != nil {
		return "", err
	}

	if rec.Verified() {
		if rec.VerifiedFor(req.MachineID) {
			return rec.Token, nil
		}
		return "", fmt.Errorf("token already used on another machine")
	}

	tkn, err := secret.Encode(req.AsUserToken())
	if err != nil {
		log.Printf("failed to encode user token error: %s", err)
		return "", internalServerError
	}
	rec.Verify(req, tkn)

	if err := r.put(rec); err != nil {
		return "", internalServerError
	}

	return tkn, nil
}

func (r *Register) notifyByEmail(email, id string) error {
	fromEmail := "ianic+org5@mantil.com"
	toEmail := email
	subject := "mantil.com sign up"

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
					Data: aws.String(fmt.Sprintf(`
Here is your activation token: %s.
Use it in you terminal to activate Mantil:

mantil activate %s
`, id, id)),
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
