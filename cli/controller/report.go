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
)

func Report(days int) error {
	msg, err := reportMessage()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return nil
		}
		return log.Wrap(err)
	}
	fs, err := newStore()
	if err != nil {
		return log.Wrap(err)
	}
	workspaceID := fs.Workspace().ID
	uploadReq := dto.UploadURLRequest{
		WorkspaceID: workspaceID,
		Message:     msg,
	}
	uploadRsp, err := backend.Report().URL(uploadReq)
	if err != nil {
		return log.Wrap(err)
	}
	if err := uploadLogs(days, uploadRsp.URL); err != nil {
		return log.Wrap(err)
	}
	confirmReq := dto.ConfirmRequest{
		ReportID: uploadRsp.ReportID,
	}
	if err := backend.Report().Confirm(confirmReq); err != nil {
		return log.Wrap(err)
	}
	ui.Info("Bug report was successfully made! We will get in touch as soon as we can on the email address you used during registration.")
	return nil
}

func reportMessage() (string, error) {
	prompt := promptui.Prompt{
		Label: "Please include an explanation with your bug report",
	}
	return prompt.Run()
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
