package external

import (
	"context"
	"os"

	"github.com/supabase-community/storage-go"
)

func UploadFile(ctx context.Context, client *storage.Client, bucket, path, localFile string) error {
	file, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = client.UploadFile(ctx, bucket, path, file)
	return err
}

func DownloadFile(ctx context.Context, client *storage.Client, bucket, path, dest string) error {
	resp, err := client.DownloadFile(ctx, bucket, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = out.ReadFrom(resp.Body)
	return err
}

func ListFiles(ctx context.Context, client *storage.Client, bucket string) ([]string, error) {
	files, err := client.ListFiles(ctx, bucket, "", nil)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, f := range files {
		names = append(names, f.Name)
	}
	return names, nil
}
