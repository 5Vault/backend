package utils

import (
	"crypto/rand"
	"encoding/base64"
	"strings"

	"github.com/oklog/ulid/v2"
)

// GenerateULID gera um ULID único em lowercase.
func GenerateULID() string {
	return strings.ToLower(ulid.Make().String())
}

// GenerateAPIKey gera uma chave de API segura em base64 URL-safe.
func GenerateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}
