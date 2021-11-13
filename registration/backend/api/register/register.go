package register

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/google/uuid"
	"github.com/mantil-io/mantil.go"
)

type Register struct{}

type DefaultRequest struct {
	Name     string
	Email    string
	Position string
	OrgSize  string
}

type DefaultResponse struct {
	ID string
}

func New() *Register {
	return &Register{}
}

const registrationsPartition = "registrations"

type Registration struct {
	ID       string
	Name     string
	Email    string
	Position string
	OrgSize  string
	Verified bool
}

func (r *Register) Default(ctx context.Context, req *DefaultRequest) (*DefaultResponse, error) {
	kv, err := newKV()
	if err != nil {
		return nil, err
	}

	id := uuid.New().String()
	reg := Registration{
		ID:       id,
		Name:     req.Name,
		Email:    req.Email,
		Position: req.Position,
		OrgSize:  req.OrgSize,
		Verified: false,
	}
	if err := kv.Put(id, reg); err != nil {
		log.Printf("kv.Put failed: %s", err)
		return nil, err
	}

	if err := r.send(reg); err != nil {
		return nil, err
	}

	return &DefaultResponse{
		ID: id,
	}, nil
}

func newKV() (*mantil.KV, error) {
	kv, err := mantil.NewKV(registrationsPartition)
	if err != nil {
		log.Printf("failed to init kv: %s", err)
	}
	return kv, err

}

func (r *Register) Verify(ctx context.Context, id string) error {
	kv, err := newKV()
	if err != nil {
		return fmt.Errorf("internal server error")
	}
	var reg Registration
	if err := kv.Get(id, &reg); err != nil {
		return fmt.Errorf("there is no registration for %s", id)
	}

	if reg.Verified {
		return nil
	}
	reg.Verified = true
	if err := kv.Put(reg.ID, reg); err != nil {
		log.Printf("kv.Put failed: %s", err)
		return fmt.Errorf("internal server error")
	}
	return nil
}

func (r *Register) Query(ctx context.Context, id string) (int, error) {
	kv, err := newKV()
	if err != nil {
		return -127, err
	}
	var reg Registration
	if err := kv.Get(id, &reg); err != nil {
		return -1, err
	}
	if !reg.Verified {
		return 0, nil
	}
	return 1, nil
}

func (r *Register) Send(ctx context.Context) error {
	fromEmail := "ianic+org5@mantil.com"
	toEmail := "igor.anic@gmail.com"
	subject := "subject of the ses message"

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
					Data: aws.String("mail content"),
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

func (r *Register) send(reg Registration) error {
	fromEmail := "ianic+org5@mantil.com"
	toEmail := reg.Email
	subject := "mantil.com registration"

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
`, reg.ID, reg.ID)),
				},
				// 				Html: &types.Content{
				// 					Data: aws.String(fmt.Sprintf(
				// 						`Click
				// <a href="https://4fc99dc1lf.execute-api.eu-central-1.amazonaws.com/register/verify?id=%s">here</a>
				// to register.`, reg.ID)),
				//},
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
