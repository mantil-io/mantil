package signup

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type UploadURLRequest struct {
	UserID  string `json:"userId"`
	Message string `json:"message"`
}

type UploadURLResponse struct {
	ReportID string `json:"reportId"`
	URL      string `json:"url"`
}

type ConfirmRequest struct {
	ReportID string `json:"reportId"`
}

type ReportRecord struct {
	ID         string
	UserID     string
	S3Key      string
	Message    string
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
		UserID:    r.UserID,
		S3Key:     fmt.Sprintf("%s/%s.zip", time.Now().Format("2006-01-02"), id),
		Message:   r.Message,
		RequestAt: time.Now().UnixMilli(),
	}
}
