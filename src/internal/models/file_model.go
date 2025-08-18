package models

type File struct {
	Url string `json:"url"`
}

type RequestFile struct {
	Data     []byte `json:"data"`
	MimeType string `json:"mime_type"`
}

type ResponseFile struct {
	ID         uint   `json:"id"`
	UserID     string `json:"user_id"`
	StorageID  string `json:"storage_id"`
	FileID     string `json:"file_id"`
	FileType   string `json:"file_type"`
	FileURL    string `json:"file_url"`
	UploadedAt string `json:"uploaded_at"`
}
