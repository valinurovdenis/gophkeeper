// Package filestorage for storing blob files.
package filestorage

import (
	"context"

	pb "github.com/valinurovdenis/gophkeeper/internal/proto"
)

// Storage contains file blobs.
//
//go:generate mockery --name StreamingFileStorage
type StreamingFileStorage interface {
	// Get file chunks.
	Download(stream pb.GophKeeperService_DownloadFileServer, fileId string) error

	// Write file chunks.
	Upload(stream pb.GophKeeperService_UploadFileServer, fileSize int64, fileId string) error

	// Delete file.
	Delete(ctx context.Context, fileId string) error
}

// Chunk size for file streaming.
const ChunkSize = 100 * 1024
