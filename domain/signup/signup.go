package signup

import (
	_ "embed"
	"net/mail"
	"strings"
	"time"

	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/token"
)

// TokenClaims content of the user token
type TokenClaims struct {
	ActivationCode string `json:"activationCode,omitempty"`
	Email          string `json:"email,omitempty"`
	WorkspaceID    string `json:"workspaceID,omitempty"`
	MachineID      string `json:"machineID,omitempty"`
	CreatedAt      int64  `json:"createdAt,omitempty"`
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
	// if ut.MachineID != domain.MachineID() {
	// 	return nil, fmt.Errorf("token not valid for this machine")
	// }
	return &ut, nil
}

// ActivateRequest data for the signup Activate method
type ActivateRequest struct {
	ID             string `json:"id,omitempty"`
	ActivationCode string `json:"activationCode,omitempty"`
	WorkspaceID    string `json:"workspaceID,omitempty"`
	MachineID      string `json:"machineID,omitempty"`
}

func (ar ActivateRequest) Code() string {
	if ar.ActivationCode != "" {
		return ar.ActivationCode
	}
	return ar.ID
}

func NewActivateRequest(activationCode, workspaceID string) ActivateRequest {
	return ActivateRequest{
		ID:             activationCode,
		ActivationCode: activationCode,
		WorkspaceID:    workspaceID,
		MachineID:      domain.MachineID(),
	}
}

func (r *ActivateRequest) Valid() bool {
	return r.ID != "" && r.MachineID != ""
}

// Record is backend database record for each user signup
type Record struct {
	ID             string
	ActivationCode string
	Email          string
	WorkspaceID    string
	MachineID      string
	CreatedAt      int64
	ActivatedAt    int64
	Token          string
	Developer      bool
	RemoteIP       string
	// survery attributes
	Name             string
	Position         string
	OrganizationSize string
	Raw              []byte
}

func (r *Record) Activate(ar ActivateRequest) {
	r.ActivationCode = ar.ActivationCode
	r.MachineID = ar.MachineID
	r.WorkspaceID = ar.WorkspaceID
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
		ActivationCode: r.ActivationCode,
		Email:          r.Email,
		WorkspaceID:    r.WorkspaceID,
		MachineID:      r.MachineID,
		CreatedAt:      time.Now().UnixMilli(),
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
		id = TestActivationCode
	}
	return Record{
		ID:               id,
		ActivationCode:   id,
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
//   * activation id is always TestActivationCode - enables test to call Activate without previously getting email
const (
	TestEmail          = "YYcdPSsHQFChMQTk0zF3Kw@mantil.com"
	TestActivationCode = "YYcdPSsHQFChMQTk0zF3Kw"
)
