package controller

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/signup"
)

var reportEndpoint = apiEndpoint{url: "https://cx0kumro6g.execute-api.eu-central-1.amazonaws.com/report"}

const (
	uploadURLEndpoint      = "url"
	notifyUploadedEndpoint = "uploaded"
)

func Report(days int) error {
	signupID, err := signupID()
	if err != nil {
		return log.Wrap(err)
	}
	uploadReq := signup.UploadURLRequest{
		SignupID: signupID,
	}
	var uploadRsp signup.UploadURLResponse
	if err := reportEndpoint.Call(uploadURLEndpoint, &uploadReq, &uploadRsp); err != nil {
		return log.Wrap(err)
	}
	if err := uploadLogs(days, uploadRsp.URL); err != nil {
		return log.Wrap(err)
	}
	uploadedReq := signup.UploadedRequest{
		ReportID: uploadRsp.ReportID,
	}
	if err := reportEndpoint.Call(notifyUploadedEndpoint, &uploadedReq, nil); err != nil {
		return log.Wrap(err)
	}
	ui.Info("Bug report was successfuly made! We will get in touch as soon as we can on the email address you used during registration.")
	return nil
}

func signupID() (string, error) {
	token, err := domain.ReadActivationToken()
	if err != nil {
		return "", log.Wrap(err)
	}
	claims, err := signup.Decode(token)
	if err != nil {
		return "", log.Wrap(err)
	}
	return claims.ID, nil
}

func uploadLogs(days int, url string) error {
	files, err := logFilesToUpload(days)
	if err != nil {
		return log.Wrap(err)
	}
	zip, err := zipFiles(files)
	if err != nil {
		return log.Wrap(err)
	}
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(zip))
	if err != nil {
		return log.Wrap(err)
	}
	rsp, err := client.Do(req)
	if err != nil {
		return log.Wrap(err)
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		return log.Wrapf("non-ok status received from logs upload: %s", rsp.Status)
	}
	return nil
}

func logFilesToUpload(days int) ([]string, error) {
	logsDir, err := log.LogsDir()
	if err != nil {
		return nil, log.Wrap(err)
	}
	var files []string
	for i := 0; i < days; i++ {
		name := log.LogFileForDate(time.Now().AddDate(0, 0, -i))
		file := filepath.Join(logsDir, name)
		// add only log files which exist
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			files = append(files, file)
		}
	}
	return files, nil
}

func zipFiles(files []string) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	for _, file := range files {
		if err := addFileToZip(w, file); err != nil {
			return nil, log.Wrap(err)
		}
	}
	if err := w.Close(); err != nil {
		return nil, log.Wrap(err)
	}
	return buf.Bytes(), nil
}

func addFileToZip(w *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return log.Wrap(err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return log.Wrap(err)
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return log.Wrap(err)
	}
	header.Method = zip.Deflate

	writer, err := w.CreateHeader(header)
	if err != nil {
		return log.Wrap(err)
	}
	_, err = io.Copy(writer, file)
	if err != nil {
		return log.Wrap(err)
	}
	return nil
}
