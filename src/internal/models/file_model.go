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
	FileSize   int64  `json:"file_size"`
}

type FileStats struct {
	TotalFiles  int64          `json:"total_files"`
	UsedSize    int64          `json:"used_size"`  // Tamanho real dos arquivos
	TotalSize   int64          `json:"total_size"` // 250 GB fixo
	FreeSpace   int64          `json:"free_space"` // Espaço livre
	RecentFiles []ResponseFile `json:"recent_files"`
}
