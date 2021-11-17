package signup

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type UploadURLRequest struct {
	SignupID string `json:"signupId"`
}

type UploadURLResponse struct {
	ReportID string `json:"reportId"`
	URL      string `json:"url"`
}

type UploadedRequest struct {
	ReportID string `json:"reportId"`
}

type ReportRecord struct {
	ID         string
	SignupID   string
	S3Key      string
	RequestAt  int64
	UploadedAt int64
}

func (r *ReportRecord) Uploaded() {
	r.UploadedAt = time.Now().UnixMilli()
}

func (r *UploadURLRequest) AsRecord() ReportRecord {
	buf := make([]byte, 22)
	uid := [16]byte(uuid.New())
	base64.RawURLEncoding.Encode(buf, uid[:])
	id := string(buf)
	return ReportRecord{
		ID:        id,
		SignupID:  r.SignupID,
		S3Key:     fmt.Sprintf("%s/%s.zip", time.Now().Format("2006-01-02"), id),
		RequestAt: time.Now().UnixMilli(),
	}
}
