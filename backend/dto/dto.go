package dto

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
