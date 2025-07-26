// Package metadatastorage for storing files metainfo.
package metadatastorage

import (
	"context"

	pb "github.com/valinurovdenis/gophkeeper/internal/proto"
)

// Storage contains file metainfo.
//
//go:generate mockery --name MetadataStorage
type MetadataStorage interface {
	// Get file info by id.
	GetFileById(context context.Context, fileId string) (*pb.FileInfo, error)

	// Get files info by login.
	GetFilesByLogin(context context.Context, login string) (*pb.ListFiles, error)

	// Add file metainfo.
	AddFileInfo(context context.Context, fileInfo *pb.FileInfo) error

	// Delete file metainfo.
	DeleteFileInfo(context context.Context, fileId string) error

	// Check whether storage alive.
	Ping() error
}
