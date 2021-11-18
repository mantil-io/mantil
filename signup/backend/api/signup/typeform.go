package signup

import (
	"encoding/base64"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mantil-io/mantil/signup"
)

type TypeformWebhook struct {
	EventID      string `json:"event_id"`
	EventType    string `json:"event_type"`
	FormResponse struct {
		FormID      string    `json:"form_id"`
		Token       string    `json:"token"`
		SubmittedAt time.Time `json:"submitted_at"`
		LandedAt    time.Time `json:"landed_at"`
		Calculated  struct {
			Score int `json:"score"`
		} `json:"calculated"`
		Variables []struct {
			Key    string `json:"key"`
			Type   string `json:"type"`
			Number int    `json:"number,omitempty"`
			Text   string `json:"text,omitempty"`
		} `json:"variables"`
		Definition struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Fields []struct {
				ID                      string `json:"id"`
				Title                   string `json:"title"`
				Type                    string `json:"type"`
				Ref                     string `json:"ref"`
				AllowMultipleSelections bool   `json:"allow_multiple_selections"`
				AllowOtherChoice        bool   `json:"allow_other_choice"`
			} `json:"fields"`
		} `json:"definition"`
		Answers []struct {
			Type  string `json:"type"`
			Text  string `json:"text,omitempty"`
			Field struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"field"`
			Email   string `json:"email,omitempty"`
			Date    string `json:"date,omitempty"`
			Choices struct {
				Labels []string `json:"labels"`
			} `json:"choices,omitempty"`
			Number  int  `json:"number,omitempty"`
			Boolean bool `json:"boolean,omitempty"`
			Choice  struct {
				Label string `json:"label"`
			} `json:"choice,omitempty"`
		} `json:"answers"`
	} `json:"form_response"`
}

func (t *TypeformWebhook) Email() string {
	for _, r := range t.FormResponse.Answers {
		if r.Field.Type == "email" {
			return r.Email
		}
	}
	return ""
}

func (t *TypeformWebhook) Valid() bool {
	_, err := mail.ParseAddress(t.Email())
	return t.Email() != "" && err == nil
}

func (r *TypeformWebhook) AsRecord() signup.SignupRecord {
	buf := make([]byte, 22)
	uid := [16]byte(uuid.New())
	base64.RawURLEncoding.Encode(buf, uid[:])
	id := string(buf)
	email := r.Email()
	return signup.SignupRecord{
		ID:        id,
		Email:     email,
		Developer: strings.HasSuffix(email, "@mantil.com"),
	}
}
