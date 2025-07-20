// Package filestorage for s3 client.
package filestorage

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/valinurovdenis/gophkeeper/internal/app/config"
)

// Client for s3.
type S3Client struct {
	client *minio.Client
}

func NewS3Client() (*S3Client, error) {
	client, err := getClient()
	if err != nil {
		return nil, fmt.Errorf("cannot create s3 client: %w", err)
	}
	return &S3Client{client: client}, nil
}

// Get minio client.
func getClient() (*minio.Client, error) {
	config := config.GetConfig()
	useSSL := false
	minioClient, errInit := minio.New(config.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.S3AccessKey, config.S3SecretKey, ""),
		Secure: useSSL,
	})
	if errInit != nil {
		log.Fatalln(errInit)
	}
	return minioClient, errInit
}

// Make minio bucket.
func (c *S3Client) MakeBucket(ctx context.Context) error {
	config := config.GetConfig()
	bucketName := config.S3Bucket
	err := c.client.MakeBucket(ctx, config.S3Bucket, minio.MakeBucketOptions{Region: config.S3Region})
	if err != nil {
		exists, errBucketExists := c.client.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			return fmt.Errorf("bucket already exists: %w", err)
		} else {
			return fmt.Errorf("error when creating bucket: %w", err)
		}
	} else {
		return nil
	}
}

// Upload file to minio.
func (c *S3Client) UploadFile(ctx context.Context, fileReader io.Reader, fileName string, fileSize int64) error {
	config := config.GetConfig()
	_, err := c.client.PutObject(ctx, config.S3Bucket, fileName, fileReader, fileSize, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	return nil
}

// Download file from minio.
func (c *S3Client) DownloadFile(ctx context.Context, fileName string) (io.Reader, error) {
	config := config.GetConfig()
	reader, err := c.client.GetObject(ctx, config.S3Bucket, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	return reader, nil
}

// Delete file from minio.
func (c *S3Client) DeleteFile(ctx context.Context, fileName string) error {
	config := config.GetConfig()
	err := c.client.RemoveObject(ctx, config.S3Bucket, fileName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
