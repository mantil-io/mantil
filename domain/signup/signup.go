package signup

import (
	_ "embed"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/token"
)

// TokenClaims content of the user token
type TokenClaims struct {
	ID        string `json:"id,omitempty"`
	Email     string `json:"email,omitempty"`
	MachineID string `json:"machineID,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
}

// Decode jwt token string to claims.
// Fails if jwt is not signed by proper private key.
func Decode(jwt, publicKey string) (TokenClaims, error) {
	var ut TokenClaims
	err := token.Decode(jwt, publicKey, &ut)
	return ut, err
}

// Validate returns true if jwt is valid for that machine
func Validate(jwt, publicKey string) (*TokenClaims, error) {
	jwt = strings.TrimSpace(jwt)
	var ut TokenClaims
	err := token.Decode(jwt, publicKey, &ut)
	if err != nil {
		return nil, err
	}
	if ut.MachineID != domain.MachineID() {
		return nil, fmt.Errorf("token not valid for this machine")
	}
	return &ut, nil
}

// ActivateRequest data for the signup Activate method
type ActivateRequest struct {
	ID        string `json:"id,omitempty"`
	MachineID string `json:"machineID,omitempty"`
}

func NewActivateRequest(id string) ActivateRequest {
	return ActivateRequest{
		ID:        id,
		MachineID: domain.MachineID(),
	}
}

func (r *ActivateRequest) Valid() bool {
	return r.ID != "" && r.MachineID != ""
}

// Record is backend database record for each user signup
type Record struct {
	ID          string
	Email       string
	MachineID   string
	CreatedAt   int64
	ActivatedAt int64
	Token       string
	Developer   bool
	RemoteIP    string
	// survery attributes
	Name             string
	Position         string
	OrganizationSize string
	Raw              []byte
}

func (r *Record) Activate(vr ActivateRequest) {
	r.MachineID = vr.MachineID
	r.ActivatedAt = time.Now().UnixMilli()
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
	Email            string `json:"email,omitempty"`
	Name             string `json:"name,omitempty"`
	Position         string `json:"position,omitempty"`
	OrganizationSize string `json:"orgSize,omitempty"`
}

// convert it to the Record
func (r *RegisterRequest) AsRecord() Record {
	id := domain.UID()
	if r.Email == TestEmail {
		id = TestID
	}
	return Record{
		ID:               id,
		Email:            r.Email,
		Name:             r.Name,
		Position:         r.Position,
		OrganizationSize: r.OrganizationSize,
		Developer:        isDeveloper(r.Email),
		CreatedAt:        time.Now().UnixMilli(),
	}
}

func isDeveloper(email string) bool {
	return strings.HasSuffix(email, "@mantil.com")
}

func (r *RegisterRequest) Valid() bool {
	_, err := mail.ParseAddress(r.Email)
	return r.Email != "" && err == nil
}

// used in backend project integration tests
// backend handles this mail specially:
//   * mail it is not sent
//   * activation id is always TestID - enables test to call Activate without previously getting email
const (
	TestEmail = "YYcdPSsHQFChMQTk0zF3Kw@mantil.com"
	TestID    = "YYcdPSsHQFChMQTk0zF3Kw"
)
