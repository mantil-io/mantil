package signup

import (
	"context"
	"fmt"
	"log"

	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil/backend/secret"
	"github.com/mantil-io/mantil/domain/signup"
)

const (
	registrationPartition = "registration"
	activationPartition   = "activation"
	workspacePartition    = "workspace"
)

var (
	internalServerError = fmt.Errorf("internal server error")
	badRequestError     = fmt.Errorf("bad request")
)

type Signup struct {
	kv *kv
}

func New() *Signup {
	return &Signup{
		kv: &kv{},
	}
}

func (r *Signup) register(ctx context.Context, req signup.RegisterRequest) (*signup.RegisterRecord, error) {
	if !req.Valid() {
		return nil, badRequestError
	}
	rec := req.ToRecord(remoteIPRawRequest(ctx))
	if err := r.kv.Registrations().Put(rec.ActivationCode, rec); err != nil {
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

func (r *Signup) activate(ctx context.Context, req signup.ActivateRequest) (*signup.ActivateRecord, *signup.RegisterRecord, error) {
	if !req.Valid() {
		return nil, nil, badRequestError
	}

	var rr signup.RegisterRecord
	if err := r.kv.Registrations().Get(req.ActivationCode, &rr); err != nil {
		log.Printf("register record not found for %s, error: %s", req.ActivationCode, err)
		return nil, nil, fmt.Errorf("activation code not found")
	}

	ar := req.ToRecord(remoteIP(ctx))
	token, err := secret.Encode(ar.ToTokenClaims())
	if err != nil {
		log.Printf("failed to encode user token error: %s", err)
		return nil, nil, internalServerError
	}
	ar.Token = token

	if err := r.kv.Activations().Put(ar.ID, ar); err != nil {
		return nil, nil, err
	}
	rr.Activations = append(rr.Activations, ar.ID)
	if err := r.kv.Registrations().Put(rr.ActivationCode, rr); err != nil {
		return nil, nil, err
	}
	wr := ar.AsWorkspaceRecord()
	if err := r.kv.Workspaces().Put(wr.ID, wr); err != nil {
		return nil, nil, err
	}

	return &ar, &rr, nil
}

func (r *Signup) Activate(ctx context.Context, req signup.ActivateRequest) (string, error) {
	ar, rr, err := r.activate(ctx, req)
	if err != nil {
		return "", err
	}

	if err := r.sendWelcomeMail(rr.Email, rr.Name); err != nil {
		log.Printf("failed to sedn welcome mail error %s", err)
		// do nothing, not critical
	}
	return ar.Token, nil
}

func (r *Signup) typeform(ctx context.Context, req signup.TypeformWebhook) (*signup.RegisterRecord, error) {
	if !req.Valid() {
		return nil, badRequestError
	}
	rec := req.AsRecord(remoteIPRawRequest(ctx))
	if err := r.kv.Registrations().Put(rec.ActivationCode, rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

// Typeform webhook for register request
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

func remoteIP(ctx context.Context) string {
	rc, ok := mantil.FromContext(ctx)
	if !ok {
		return ""
	}
	return rc.Request.RemoteIP()
}

func remoteIPRawRequest(ctx context.Context) (string, []byte) {
	rc, ok := mantil.FromContext(ctx)
	if !ok {
		return "", nil
	}
	return rc.Request.RemoteIP(), rc.Request.Raw
}
