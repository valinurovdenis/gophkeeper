// Package service contains methods for file processing.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/valinurovdenis/gophkeeper/internal/app/encryption"
	"github.com/valinurovdenis/gophkeeper/internal/app/filestorage"
	"github.com/valinurovdenis/gophkeeper/internal/app/metadatastorage"
	pb "github.com/valinurovdenis/gophkeeper/internal/proto"
)

// Error in case when user with given login doesnt own file with given id.
var ErrNotOwn = errors.New("file not owned")

type GophKeeperService struct {
	fileStorage     filestorage.StreamingFileStorage
	metaDataStorage metadatastorage.MetadataStorage
}

func NewGophKeeperService(s3Storage filestorage.StreamingFileStorage, metaDataStorage metadatastorage.MetadataStorage) (*GophKeeperService, error) {
	return &GophKeeperService{fileStorage: s3Storage, metaDataStorage: metaDataStorage}, nil
}

func (h *GophKeeperService) GetUserFiles(ctx context.Context, login string) (*pb.ListFiles, error) {
	files, err := h.metaDataStorage.GetFilesByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("error getting file metainfo: %w", err)
	}
	return files, nil
}

func (h *GophKeeperService) UploadFile(stream pb.GophKeeperService_UploadFileServer, login string) error {
	res, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}
	info := res.GetInfo()
	if info == nil {
		return fmt.Errorf("no upload file info")
	}
	_, err = encryption.DecryptFileEncryptionKey(info.GetEncryptionKey(), encryption.ServerPrivateKey())
	if err != nil {
		return fmt.Errorf("wrong encryption key %w", err)
	}

	info.Login = login
	if info.GetCreated() == 0 {
		info.Created = uint64(time.Now().Unix())
	}
	info.Id = &pb.FileId{Id: filestorage.CreateFileId(info)}

	fileSize := int64(info.GetSize())
	err = h.fileStorage.Upload(stream, fileSize, info.GetId().Id)
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}
	stream.SendAndClose(&pb.UploadResponse{Id: &pb.FileId{Id: info.GetId().Id}})
	return h.metaDataStorage.AddFileInfo(stream.Context(), info)
}

func (h *GophKeeperService) DownloadFile(fileId *pb.FileId, stream pb.GophKeeperService_DownloadFileServer, login string, clientPublicKey []byte) error {
	info, err := h.metaDataStorage.GetFileById(stream.Context(), fileId.GetId())
	if err != nil {
		return fmt.Errorf("error getting file metainfo: %w", err)
	}
	if info.Login != login {
		return ErrNotOwn
	}
	key, _ := encryption.DecryptFileEncryptionKey(info.EncryptionKey, encryption.ServerPrivateKey())
	encryptedKey, err := encryption.EncryptFileEncryptionKey(key, clientPublicKey)
	if err != nil {
		return fmt.Errorf("cannot encrypt file encryption key: %w", err)
	}
	stream.Send(&pb.FileStream{Data: &pb.FileStream_Info{Info: &pb.FileInfo{
		Id:            info.Id,
		Filename:      info.Filename,
		Login:         info.Login,
		Comment:       info.Comment,
		Created:       info.Created,
		Size:          info.Size,
		EncryptionKey: encryptedKey}}})
	return h.fileStorage.Download(stream, fileId.GetId())
}

func (h *GophKeeperService) DeleteFile(ctx context.Context, fileId *pb.FileId, login string) error {
	info, err := h.metaDataStorage.GetFileById(ctx, fileId.GetId())
	if err != nil {
		return fmt.Errorf("error getting file metainfo: %w", err)
	}
	if info.Login != login {
		return ErrNotOwn
	}
	return h.fileStorage.Delete(ctx, fileId.GetId())
}
