package report

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil/signup"
)

var (
	internalServerError = fmt.Errorf("internal server error")
)

const (
	SlackWebhookEnv = "SLACK_WEBHOOK"
	ReportBucketEnv = "REPORT_BUCKET"

	reportsPartition = "reports"
)

type Report struct {
	kv *mantil.KV
}

func New() *Report {
	return &Report{}
}

func (r *Report) URL(ctx context.Context, req signup.UploadURLRequest) (*signup.UploadURLResponse, error) {
	bucket, ok := os.LookupEnv(ReportBucketEnv)
	if !ok {
		log.Printf("report bucket env %s not set", ReportBucketEnv)
		return nil, internalServerError
	}
	rec := req.AsRecord()
	url, err := s3PutURL(bucket, rec.S3Key)
	if err != nil {
		log.Printf("s3 put url failed: %s", err)
		return nil, internalServerError
	}
	if err := r.put(rec); err != nil {
		log.Printf("report put failed: %s", err)
		return nil, internalServerError
	}
	return &signup.UploadURLResponse{
		ReportID: rec.ID,
		URL:      url,
	}, nil
}

func (r *Report) Uploaded(ctx context.Context, req signup.UploadedRequest) error {
	rec, err := r.get(req.ReportID)
	if err != nil {
		return fmt.Errorf("kv.get failed: %s", err)
	}
	rec.Uploaded()
	if err := r.put(rec); err != nil {
		return fmt.Errorf("kv.put failed: %s", err)
	}
	webhook, ok := os.LookupEnv(SlackWebhookEnv)
	if !ok {
		log.Printf("slack webhook env %s not set", SlackWebhookEnv)
		return internalServerError
	}
	msg := notificationMessage(rec)
	if err := r.notifyToSlack(webhook, msg); err != nil {
		log.Printf("notify to slack: %s", err)
		return internalServerError
	}
	return nil
}

func notificationMessage(rec signup.ReportRecord) string {
	msg := fmt.Sprintf("New bug report with id %s uploaded!", rec.ID)
	bucket, ok := os.LookupEnv(ReportBucketEnv)
	if !ok {
		log.Printf("report bucket env not set while creating notification message")
		return msg
	}
	url, err := s3GetURL(bucket, rec.S3Key)
	if err != nil {
		log.Printf("s3 get url failed: %s", err)
		return msg
	}
	dwl := fmt.Sprintf("Logs of the report can be downloaded <%s|here>", url)
	return fmt.Sprintf("%s\n%s", msg, dwl)
}

func (r *Report) notifyToSlack(url string, text string) error {
	msg := struct {
		Text string `json:"text"`
	}{
		text,
	}
	buf, _ := json.Marshal(msg)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		return fmt.Errorf("new request failed: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %s", err)
	}
	respb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response failed: %s", err)
	}
	if string(respb) != "ok" {
		return fmt.Errorf("non-ok response returned: %s", string(respb))
	}
	return nil
}

func (r *Report) connectKV() error {
	if r.kv != nil {
		return nil
	}

	kv, err := mantil.NewKV(reportsPartition)
	if err != nil {
		return err
	}
	r.kv = kv
	return nil
}

func (r *Report) put(rec signup.ReportRecord) error {
	if err := r.connectKV(); err != nil {
		return err
	}

	if err := r.kv.Put(rec.ID, rec); err != nil {
		return err
	}
	return nil
}

func (r *Report) get(id string) (signup.ReportRecord, error) {
	var rec signup.ReportRecord

	if err := r.connectKV(); err != nil {
		return rec, err
	}

	if err := r.kv.Get(id, &rec); err != nil {
		return rec, err
	}

	return rec, nil
}

func s3PutURL(bucket, key string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return "", err
	}
	client := s3.NewPresignClient(s3.NewFromConfig(cfg))

	poi := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	pou, err := client.PresignPutObject(context.Background(), poi)
	if err != nil {
		return "", err
	}
	return pou.URL, nil
}

func s3GetURL(bucket, key string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return "", err
	}
	client := s3.NewPresignClient(s3.NewFromConfig(cfg))

	goi := &s3.GetObjectInput{
		Bucket:          aws.String(bucket),
		Key:             aws.String(key),
		ResponseExpires: aws.Time(time.Now().Add(12 * time.Hour)),
	}
	gou, err := client.PresignGetObject(context.Background(), goi)
	if err != nil {
		return "", err
	}
	return gou.URL, nil
}
