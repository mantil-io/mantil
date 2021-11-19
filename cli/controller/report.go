package controller

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/backend/dto"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/signup"
)

var reportEndpoint = apiEndpoint{url: "https://ytg5gfkg5k.execute-api.eu-central-1.amazonaws.com/report"}

const (
	uploadURLEndpoint     = "url"
	confirmUploadEndpoint = "confirm"
)

func Report(days int) error {
	msg, err := reportMessage()
	if err != nil {
		ui.Info("Submitting report aborted.")
		return nil
	}
	userID, err := userID()
	if err != nil {
		return log.Wrap(err)
	}
	uploadReq := dto.UploadURLRequest{
		UserID:  userID,
		Message: msg,
	}
	var uploadRsp dto.UploadURLResponse
	if err := reportEndpoint.Call(uploadURLEndpoint, &uploadReq, &uploadRsp); err != nil {
		return log.Wrap(err)
	}
	if err := uploadLogs(days, uploadRsp.URL); err != nil {
		return log.Wrap(err)
	}
	confirmReq := dto.ConfirmRequest{
		ReportID: uploadRsp.ReportID,
	}
	if err := reportEndpoint.Call(confirmUploadEndpoint, &confirmReq, nil); err != nil {
		return log.Wrap(err)
	}
	ui.Info("Bug report was successfully made! We will get in touch as soon as we can on the email address you used during registration.")
	return nil
}

func reportMessage() (string, error) {
	prompt := promptui.Prompt{
		Label: "Please include an explanation with your bug report",
	}
	res, err := prompt.Run()
	if err != nil {
		return "", err
	}
	return res, nil
}

func userID() (string, error) {
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
