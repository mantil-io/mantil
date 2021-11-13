package registration

import (
	_ "embed"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/mantil-io/mantil/kit/token"
)

//go:embed public_key
var publicKey string

type UserToken struct {
	ID        string `json:"id,omitempty"`
	Email     string `json:"email,omitempty"`
	MachineID string `json:"machineID,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
}

func (u *UserToken) Verified() bool {
	return u.MachineID != ""
}

func (u *UserToken) VerifiedWith(req VerifyRequest) bool {
	return u.MachineID == req.MachineID
}

func Decode(tkn string) (UserToken, error) {
	var ut UserToken
	err := token.Decode(tkn, publicKey, &ut)
	return ut, err
}

type VerifyRequest struct {
	ID        string `json:"id,omitempty"`
	Email     string `json:"email,omitempty"`
	MachineID string `json:"machineID,omitempty"`
}

func (r *VerifyRequest) Valid() bool {
	return r.ID != "" && r.MachineID != "" && r.Email != ""
}

func (r *VerifyRequest) AsRecord() Record {
	return Record{
		ID:        r.ID,
		Email:     r.Email,
		MachineID: r.MachineID,
		CreatedAt: time.Now().UnixMilli(),
	}
}

func (r *VerifyRequest) AsUserToken() UserToken {
	return UserToken{
		ID:        r.ID,
		Email:     r.Email,
		MachineID: r.MachineID,
		CreatedAt: time.Now().UnixMilli(),
	}
}

type Record struct {
	ID         string
	Email      string
	MachineID  string
	CreatedAt  int64
	VerifiedAt int64
	Token      string
	Survey     Survey
}

type Survey struct {
	Name     string
	Position string
	OrgSize  string
}

func (r *Record) Verify(vr VerifyRequest, tkn string) {
	r.MachineID = vr.MachineID
	r.VerifiedAt = time.Now().UnixMilli()
	r.Token = tkn
}

func (r *Record) Verified() bool {
	return r.Token != ""
}

func (r *Record) VerifiedFor(machineID string) bool {
	return r.MachineID == machineID
}

type RegisterRequest struct {
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
	Position string `json:"position,omitempty"`
	OrgSize  string `json:"orgSize,omitempty"`
}

func (r *RegisterRequest) AsRecord() Record {
	id := uuid.New()
	s := Survey{
		Position: r.Position,
		OrgSize:  r.OrgSize,
	}
	return Record{
		ID:        id.String(),
		Email:     r.Email,
		Survey:    s,
		CreatedAt: time.Now().UnixMilli(),
	}
}

func (r *RegisterRequest) Valid() bool {
	_, err := mail.ParseAddress(r.Email)
	return r.Email != "" && err == nil
}
