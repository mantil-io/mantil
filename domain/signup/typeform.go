package signup

import (
	"net/mail"
	"time"

	"github.com/mantil-io/mantil/domain"
)

// This data structure is build with example json payload from typeform site:
// https://developer.typeform.com/webhooks/example-payload/
// which I pase into json to Go struct generator:
// https://mholt.github.io/json-to-go/
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

func (r *TypeformWebhook) AsRecord(ip string, raw []byte) RegisterRecord {
	id := domain.UID()
	email := r.Email()
	return RegisterRecord{
		ActivationCode:   id,
		Email:            email,
		Name:             r.AnswerByID("35xdSkzCv9q9"),
		Position:         r.AnswerByID("C4PHTxIvSRYg"),
		OrganizationSize: r.AnswerByID("9jdxqysanTG9"),
		Developer:        isDeveloper(email),
		CreatedAt:        time.Now().UnixMilli(),
		RemoteIP:         ip,
		Raw:              raw,
		Source:           SourceTypeform,
		Survey:           r.Survey(),
	}
}

func (t TypeformWebhook) Answer(no int) string {
	if len(t.FormResponse.Answers) > no {
		a := t.FormResponse.Answers[no]
		if a.Text != "" {
			return a.Text
		}
		if a.Choice.Label != "" {
			return a.Choice.Label
		}
		if a.Email != "" {
			return a.Email
		}
	}
	return ""
}

func (t TypeformWebhook) AnswerByID(id string) string {
	for no, a := range t.FormResponse.Answers {
		if a.Field.ID == id {
			return t.Answer(no)
		}
	}
	return ""
}

func (t TypeformWebhook) Survey() map[string]string {
	m := make(map[string]string)
	for _, q := range t.FormResponse.Definition.Fields {
		m[q.Title] = t.AnswerByID(q.ID)
	}
	return m
}
