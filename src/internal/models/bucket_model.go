package models

// ── Bucket ────────────────────────────────────────────────────────────────────

type RequestCreateBucket struct {
	Name string `json:"name" binding:"required,min=1,max=64"`
}

type ResponseBucket struct {
	BucketID            string `json:"bucket_id"`
	UserID              string `json:"user_id"`
	Name                string `json:"name"`
	R2Name              string `json:"r2_name"`
	Status              string `json:"status"`
	CustomDomain        string `json:"custom_domain"`
	PublicDomain        string `json:"public_domain"`
	PublicAccessEnabled bool   `json:"public_access_enabled"`
	CreatedAt           string `json:"created_at"`
}

type RequestSetDomain struct {
	Domain string `json:"domain" binding:"required"`
}

// ── Diretório ─────────────────────────────────────────────────────────────────

type RequestCreateDirectory struct {
	Name string `json:"name" binding:"required,min=1,max=64"`
}

type ResponseDirectory struct {
	DirID     string `json:"dir_id"`
	BucketID  string `json:"bucket_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// ── Stats ─────────────────────────────────────────────────────────────────────

type BucketStats struct {
	TotalFiles int64 `json:"total_files"`
	BytesUsed  int64 `json:"bytes_used"`
}

// ── Arquivos ──────────────────────────────────────────────────────────────────

type ResponseUploadFile struct {
	FileName  string `json:"file_name"`
	PublicURL string `json:"public_url"`
	Size      int64  `json:"size"`
}

type ResponseListFiles struct {
	Files      []FileEntry `json:"files"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

type FileEntry struct {
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified"`
	PublicURL    string `json:"public_url"`
}
