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
	ActivationID   string `json:"activationID,omitempty"`
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

// RegisterRequest from cli
type RegisterRequest struct {
	Email            string `json:"email,omitempty"`
	Name             string `json:"name,omitempty"`
	Position         string `json:"position,omitempty"`
	OrganizationSize string `json:"orgSize,omitempty"`
}

// Valid if email looks good
func (r *RegisterRequest) Valid() bool {
	_, err := mail.ParseAddress(r.Email)
	return r.Email != "" && err == nil
}

// RegisterRequest is database record for a registration
type RegisterRecord struct {
	ActivationCode string   // primary key, each register request gets new activation code
	Activations    []string // activation made with this activation token, link to ActivateRecord
	Developer      bool     // from email
	RemoteIP       string   // of the request
	CreatedAt      int64    // unix milli
	Source         int      // cli or typeform
	// survery attributes
	Email            string
	Name             string
	Position         string
	OrganizationSize string
	Raw              []byte // raw http request
	Survey           map[string]string
}

const (
	SourceCli      = 1
	SourceTypeform = 2
)

// ToRecord creates Record from Request
func (r *RegisterRequest) ToRecord(ip string, raw []byte) RegisterRecord {
	code := domain.UID()      // new activation code
	if r.Email == TestEmail { // just for integration tests
		code = TestActivationCode
	}
	return RegisterRecord{
		ActivationCode:   code,
		Email:            r.Email,
		Name:             r.Name,
		Position:         r.Position,
		OrganizationSize: r.OrganizationSize,
		Developer:        isDeveloper(r.Email),
		CreatedAt:        time.Now().UnixMilli(),
		RemoteIP:         ip,
		Raw:              raw,
		Source:           SourceCli,
	}
}

// ActivateRequest data from cli activate method
type ActivateRequest struct {
	ActivationCode string `json:"activationCode,omitempty"`
	WorkspaceID    string `json:"workspaceID,omitempty"`
	MachineID      string `json:"machineID,omitempty"`
}

// NewActivateRequest used in cli to create new request
func NewActivateRequest(activationCode, workspaceID string) ActivateRequest {
	return ActivateRequest{
		ActivationCode: activationCode,
		WorkspaceID:    workspaceID,
		MachineID:      domain.MachineID(),
	}
}

// Valid only if we have all attributes
func (r *ActivateRequest) Valid() bool {
	return r.ActivationCode != "" && r.MachineID != "" && r.WorkspaceID != ""
}

// ToRecord trasforms Request to database Record
func (r *ActivateRequest) ToRecord(remoteIP string) ActivateRecord {
	return ActivateRecord{
		ID:             domain.UID(),
		ActivationCode: r.ActivationCode,
		WorkspaceID:    r.WorkspaceID,
		MachineID:      r.MachineID,
		RemoteIP:       remoteIP,
		CreatedAt:      time.Now().UnixMilli(),
	}
}

type ActivateRecord struct {
	ID             string // every activation has unique id
	ActivationCode string // link to registration
	WorkspaceID    string // from cli
	MachineID      string // from cli
	Token          string // generated for this activation
	RemoteIP       string // from where we got request
	CreatedAt      int64
}

func (r ActivateRecord) ToTokenClaims() TokenClaims {
	return TokenClaims{
		ActivationCode: r.ActivationCode,
		ActivationID:   r.ID,
		WorkspaceID:    r.WorkspaceID,
		MachineID:      r.MachineID,
		CreatedAt:      time.Now().UnixMilli(),
	}
}

func (r ActivateRecord) AsWorkspaceRecord() WorkspaceRecord {
	return WorkspaceRecord{
		ID:             r.WorkspaceID,
		ActivationCode: r.ActivationCode,
		ActivationID:   r.ID,
		CreatedAt:      time.Now().UnixMilli(),
	}
}

type WorkspaceRecord struct {
	ID             string // workspace id
	ActivationCode string // link to registration
	ActivationID   string // link to activation
	CreatedAt      int64
}

// enogh trivial to be removed?
func isDeveloper(email string) bool {
	return strings.HasSuffix(email, "@mantil.com")
}

// used in backend project integration tests
// backend handles this mail specially:
//   * mail it is not sent
//   * activation id is always TestActivationCode - enables test to call Activate without previously getting email
const (
	TestEmail          = "YYcdPSsHQFChMQTk0zF3Kw@mantil.com"
	TestActivationCode = "YYcdPSsHQFChMQTk0zF3Kw"
)
