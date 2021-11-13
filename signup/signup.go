package signup

import (
	_ "embed"
	"encoding/base64"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mantil-io/mantil/kit/token"
)

//go:embed public_key
var publicKey string

// TokenClaims content of the user token
type TokenClaims struct {
	ID        string `json:"id,omitempty"`
	Email     string `json:"email,omitempty"`
	MachineID string `json:"machineID,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
}

// Decode jwt token string to claims.
// Fails if jwt is not signed by proper private key.
func Decode(jwt string) (TokenClaims, error) {
	var ut TokenClaims
	err := token.Decode(jwt, publicKey, &ut)
	return ut, err
}

// IsValidToken returns true if jwt is valid for that machine
func IsValidToken(jwt, machineID string) bool {
	jwt = strings.TrimSpace(jwt)
	var ut TokenClaims
	err := token.Decode(jwt, publicKey, &ut)
	if err != nil {
		return false
	}
	return ut.MachineID == machineID
}

// ActivateRequest data for the signup Activate method
type ActivateRequest struct {
	ID        string `json:"id,omitempty"`
	MachineID string `json:"machineID,omitempty"`
}

func (r *ActivateRequest) Valid() bool {
	return r.ID != "" && r.MachineID != ""
}

// Record is backend database record for each user signup
type Record struct {
	ID         string
	Email      string
	MachineID  string
	CreatedAt  int64
	VerifiedAt int64
	Token      string
	Survey     Survey
}

// Survey user responses
type Survey struct {
	Name     string
	Position string
	OrgSize  string
}

func (r *Record) Activate(vr ActivateRequest) {
	r.MachineID = vr.MachineID
	r.VerifiedAt = time.Now().UnixMilli()
}

func (r *Record) Activated() bool {
	return r.Token != ""
}

func (r *Record) ActivatedFor(machineID string) bool {
	return r.MachineID == machineID
}

func (r *Record) AsTokenClaims() TokenClaims {
	return TokenClaims{
		ID:        r.ID,
		Email:     r.Email,
		MachineID: r.MachineID,
		CreatedAt: time.Now().UnixMilli(),
	}
}

// RegisterRequest data for signup Register method
type RegisterRequest struct {
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
	Position string `json:"position,omitempty"`
	OrgSize  string `json:"orgSize,omitempty"`
}

// convert it to the Record
func (r *RegisterRequest) AsRecord() Record {
	buf := make([]byte, 22)
	uid := [16]byte(uuid.New())
	base64.RawURLEncoding.Encode(buf, uid[:])
	s := Survey{
		Name:     r.Name,
		Position: r.Position,
		OrgSize:  r.OrgSize,
	}
	return Record{
		ID:        string(buf),
		Email:     r.Email,
		Survey:    s,
		CreatedAt: time.Now().UnixMilli(),
	}
}

func (r *RegisterRequest) Valid() bool {
	_, err := mail.ParseAddress(r.Email)
	return r.Email != "" && err == nil
}
