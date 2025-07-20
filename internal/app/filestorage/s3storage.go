// Package filestorage for storing blob files.
package filestorage

import (
	"context"
	"fmt"

	pb "github.com/valinurovdenis/gophkeeper/internal/proto"
)

// Storage contains s3 file blobs.
type S3FileStorage struct {
	client *S3Client
}

func NewS3FileStorage() (*S3FileStorage, error) {
	cl, err := NewS3Client()
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 client: %w", err)
	}
	return &S3FileStorage{client: cl}, nil
}

func (s *S3FileStorage) Download(stream pb.GophKeeperService_DownloadFileServer, fileId string) error {
	reader, err := s.client.DownloadFile(stream.Context(), fileId)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	return FromReader2FileStream(reader, stream)
}

func (s *S3FileStorage) Upload(stream pb.GophKeeperService_UploadFileServer, fileSize int64, fileId string) error {
	reader := NewFileStreamReader(stream)
	return s.client.UploadFile(stream.Context(), reader, fileId, fileSize)
}

func (s *S3FileStorage) Delete(ctx context.Context, fileId string) error {
	return s.client.DeleteFile(ctx, fileId)
}
