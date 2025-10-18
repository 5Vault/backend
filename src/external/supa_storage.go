package external

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/supabase-community/storage-go"
)

type SupaStorage struct{}

func NewSupaStorage() *SupaStorage {
	return &SupaStorage{}
}

var apiKey = os.Getenv("SUPABASE_API")
var url = "https://" + os.Getenv("SUPABASE_ID") + ".supabase.co/storage/v1"
var storageClient = storage_go.NewClient(url, apiKey, nil)

func (s *SupaStorage) UploadFile(userID string, localFile []byte, typeFile string, fileName string) (*storage_go.FileUploadResponse, error) {

	_, err := storageClient.GetBucket(userID)
	if err != nil {

		_, err := storageClient.CreateBucket(userID, storage_go.BucketOptions{
			Public: true,
		})
		if err != nil {
			return nil, fmt.Errorf("erro ao criar bucket: %w", err)
		}
	}

	reader := bytes.NewReader(localFile)

	res, err := storageClient.UploadFile(userID, fileName, reader, storage_go.FileOptions{
		ContentType: &typeFile,
	})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *SupaStorage) ListFiles(userID string) ([]string, error) {
	files, err := storageClient.ListFiles(userID, "/", storage_go.FileSearchOptions{})
	if err != nil {
		return nil, err
	}
	var names []string
	for _, f := range files {
		names = append(names, f.Name)
	}
	return names, nil
}

func (s *SupaStorage) GetBucket(bucketId string) (*storage_go.Bucket, error) {
	bucket, err := storageClient.GetBucket(bucketId)
	if err != nil {
		return nil, err
	}
	return &bucket, nil
}

func (s *SupaStorage) CreateBucket(bucketId string) error {
	_, err := storageClient.CreateBucket(bucketId, storage_go.BucketOptions{
		Public: true,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *SupaStorage) DownloadFile(fileURL string) ([]byte, error) {
	// Faz requisição HTTP para buscar o arquivo do storage
	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição para o storage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao baixar arquivo: status %d", resp.StatusCode)
	}

	// Lê o conteúdo do arquivo
	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler conteúdo do arquivo: %w", err)
	}

	return fileData, nil
}
