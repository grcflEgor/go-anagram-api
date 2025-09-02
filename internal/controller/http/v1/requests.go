package v1

// GroupRequest представляет запрос на группировку анаграмм
type GroupRequest struct {
	// Список слов для группировки
	Words []string `json:"words" validate:"min=1,dive,required" example:"[\"cat\",\"act\",\"tac\"]"`
	// Учитывать ли регистр при группировке
	CaseSensitive bool `json:"case_sensitive" example:"false"`
}

// UploadRequest представляет запрос на загрузку файла
type UploadRequest struct {
	// Имя файла
	Filename string `json:"filename" validate:"required" example:"words.txt"`
	// Содержимое файла
	FileContent []byte `json:"file_content" validate:"required"`
	// Учитывать ли регистр при группировке
	CaseSensitive bool `json:"case_sensitive" example:"false"`
}
