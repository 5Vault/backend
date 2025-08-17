package models

type File struct {
	Url string `json:"url"`
}

type RequestFile struct {
	Data     []byte `json:"data"`
	MimeType string `json:"mime_type"`
}
