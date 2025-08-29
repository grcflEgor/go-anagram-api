package v1

type GroupRequest struct {
	Words         []string `json:"words" validate:"min=1,dive,required"`
	CaseSensitive bool     `json:"case_sensitive"`
}

type UploadRequest struct {
	Filename      string `json:"filename" validate:"required"`
	FileContent   []byte `json:"file_content" validate:"required"`
	CaseSensitive bool   `json:"case_sensitive"`
}
