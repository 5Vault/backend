package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// GenerateRandomID gera um ID aleatório de 10 caracteres hexadecimais
func GenerateRandomID() string {
	bytes := make([]byte, 5) // 5 bytes = 10 caracteres hex
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

// GenerateAPIKey gera uma chave de API segura
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 256 bits de entropia
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Codifica em base64 URL-safe (sem padding)
	apiKey := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes)
	return apiKey, nil
}

// GenerateAPIKeyWithPrefix gera uma chave de API com prefixo personalizado
func GenerateAPIKeyWithPrefix(prefix string) (string, error) {
	bytes := make([]byte, 24) // 24 bytes para manter tamanho razoável com prefixo
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	keyPart := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes)
	return fmt.Sprintf("%s_%s", prefix, keyPart), nil
}

// GenerateHexAPIKey gera uma chave de API em formato hexadecimal
func GenerateHexAPIKey(length int) (string, error) {
	if length <= 0 {
		length = 32 // valor padrão
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
