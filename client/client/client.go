// Package client for sending requests from client.
package client

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"github.com/valinurovdenis/gophkeeper/internal/app/config"
	"github.com/valinurovdenis/gophkeeper/internal/app/encryption"
	"github.com/valinurovdenis/gophkeeper/internal/app/filestorage"
	pb "github.com/valinurovdenis/gophkeeper/internal/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type GophKeeperClient struct {
	client pb.GophKeeperServiceClient
}

func NewGophKeeperClient() (*GophKeeperClient, error) {
	conn, err := grpc.NewClient(":8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("cannot create client: %w", err)
	}
	return &GophKeeperClient{client: pb.NewGophKeeperServiceClient(conn)}, nil
}

func AddAuthTokenToContext(ctx context.Context) (context.Context, error) {
	config := config.GetConfig()
	var token string
	if file, err := os.OpenFile(config.AuthTokenFile, os.O_RDONLY, 0644); err == nil {
		defer file.Close()
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, file)
		if err != nil {
			return nil, fmt.Errorf("cannot read token from file %w", err)
		}
		token = buf.String()
	}
	md := metadata.Pairs("Authorization", token)
	newCtx := metadata.NewOutgoingContext(ctx, md)
	return newCtx, nil
}

func prettifySize(size uint64) string {
	sizef := float64(size)
	for _, unit := range []string{"", "KB", "MB", "GB"} {
		if math.Abs(sizef) < 1024.0 {
			return fmt.Sprintf("%3.1f%s", sizef, unit)
		}
		sizef /= 1024.0
	}
	return fmt.Sprintf("%.1fTB", sizef)
}

func paramIsEmpty(param string, name string) bool {
	if param == "" {
		fmt.Printf("%s must be not empty\n", name)
		return true
	}
	return false
}

func (c *GophKeeperClient) uploadFileWithProgress(stream pb.GophKeeperService_UploadFileClient, file *os.File, totalSize uint64, filename string, comment string) {
	key, err := encryption.GenerateSymmetricFileEncryptionKey()
	if err != nil {
		fmt.Println(err)
		return
	}
	encryptedKey, err := encryption.EncryptFileEncryptionKey(key, encryption.ServerPublicKey())
	if err != nil {
		fmt.Println(err)
		return
	}
	stream.Send(&pb.FileStream{Data: &pb.FileStream_Info{Info: &pb.FileInfo{Filename: filename, Comment: comment, Size: totalSize, EncryptionKey: encryptedKey}}})
	reader := bufio.NewReader(file)
	buffer := make([]byte, filestorage.ChunkSize)
	uploadedSize := int64(0)
	progressBar := NewProgressBar("Uploading", int64(totalSize))
	for {
		progressBar.Set(uploadedSize)
		n, errProgress := reader.Read(buffer)
		if errProgress != nil {
			if errProgress == io.EOF {
				break
			}
			fmt.Println(errProgress)
			return
		}
		encryptedBuf, errProgress := encryption.EncryptFileData(key, buffer[:n])
		if errProgress != nil {
			fmt.Printf("Failed to encrypt data: %s\n", errProgress)
			return
		}
		if errProgress = stream.Send(&pb.FileStream{Data: &pb.FileStream_ChunkData{ChunkData: encryptedBuf}}); err != nil {
			fmt.Println(errProgress)
			return
		}
		uploadedSize += int64(n)
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Println(err)
		return
	}
	progressBar.End()
	fmt.Println("Successfully uploaded, file id: ", resp.GetId().Id)
}

func (c *GophKeeperClient) UploadFile(ctx context.Context, filePath string, filename string, comment string) {
	if paramIsEmpty(filePath, "path") {
		return
	}
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("cannot get file info: %s\n", err)
		return
	}

	ctx, err = AddAuthTokenToContext(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	stream, err := c.client.UploadFile(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	c.uploadFileWithProgress(stream, file, uint64(fileInfo.Size()), filename, comment)
}

func (c *GophKeeperClient) downloadFileWithProgress(ctx context.Context, fileId string, file *os.File) {
	stream, err := c.client.DownloadFile(ctx, &pb.FileId{Id: fileId})
	if err != nil {
		fmt.Println(err)
		return
	}
	uploadedSize := int64(0)

	res, err := stream.Recv()
	if err != nil || res.GetInfo() == nil {
		fmt.Println("Can't get file metainfo.")
		return
	}
	encryptedKey, err := encryption.DecryptFileEncryptionKey(res.GetInfo().GetEncryptionKey(), encryption.ClientPrivateKey())
	if err != nil {
		fmt.Println("can't decrypt encryption key from file metainfo.")
		return
	}
	progressBar := NewProgressBar("Downloading", int64(res.GetInfo().Size))
	for {
		progressBar.Set(uploadedSize)
		res, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Println(err)
			return
		}
		if len(res.GetChunkData()) > 0 {
			decryptedData, err := encryption.DecryptFileData(encryptedKey, res.GetChunkData())
			if err != nil {
				fmt.Println("error when decrypt file data.")
				return
			}
			file.Write(decryptedData)
			uploadedSize += int64(len(res.GetChunkData()))
		}
	}
	progressBar.End()
}

func (c *GophKeeperClient) DownloadFile(ctx context.Context, filePath string, fileId string) {
	if paramIsEmpty(filePath, "path") || paramIsEmpty(fileId, "id") {
		return
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	ctx, err = AddAuthTokenToContext(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	c.downloadFileWithProgress(ctx, fileId, file)
}

func (c *GophKeeperClient) DeleteFile(ctx context.Context, fileId string) {
	if paramIsEmpty(fileId, "id") {
		return
	}
	ctx, err := AddAuthTokenToContext(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = c.client.DeleteFile(ctx, &pb.FileId{Id: fileId})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("File has been deleted")
}

func saveAuthToken(header metadata.MD) error {
	if len(header.Get("Authorization")) == 0 {
		return fmt.Errorf("empty authorization header")
	}
	token := header.Get("Authorization")[0]

	var err error
	var file *os.File
	if file, err = os.OpenFile(config.GetConfig().AuthTokenFile, os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		defer file.Close()
		_, err = file.WriteString(token)
	}
	return err
}

func (c *GophKeeperClient) Register(ctx context.Context, login string, password string) {
	if paramIsEmpty(login, "login") || paramIsEmpty(password, "password") {
		return
	}
	encryption.CreateKeysIfAbsent(false)
	var header metadata.MD
	var err error
	var serverPublicKey *pb.ServicePublicKey
	if serverPublicKey, err = c.client.Register(ctx, &pb.UserData{Login: login, Password: password, PublicKey: encryption.ClientPublicKey()}, grpc.Header(&header)); err == nil {
		if err = encryption.SaveKeyToFile(serverPublicKey.GetPublicKey(), config.GetConfig().ServerPublicKeyPath); err == nil {
			err = saveAuthToken(header)
		}
	}
	if err != nil {
		fmt.Println(err)
	}
}

func (c *GophKeeperClient) Login(ctx context.Context, login string, password string) {
	if paramIsEmpty(login, "login") || paramIsEmpty(password, "password") {
		return
	}
	encryption.CreateKeysIfAbsent(false)
	var header metadata.MD
	var err error
	var serverPublicKey *pb.ServicePublicKey
	if serverPublicKey, err = c.client.Login(ctx, &pb.UserData{Login: login, Password: password, PublicKey: encryption.ClientPublicKey()}, grpc.Header(&header)); err == nil {
		if err = encryption.SaveKeyToFile(serverPublicKey.GetPublicKey(), config.GetConfig().ServerPublicKeyPath); err == nil {
			err = saveAuthToken(header)
		}
	}
	if err != nil {
		fmt.Println(err)
	}
}

func (c *GophKeeperClient) ListFiles(ctx context.Context) {
	ctx, err := AddAuthTokenToContext(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	listFiles, err := c.client.GetUserFiles(ctx, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(listFiles.Files) == 0 {
		fmt.Println("No files")
	}
	for _, val := range listFiles.Files {
		created := time.Unix(int64(val.Created), 0)
		fmt.Printf("id=%s    filename='%s'    created=%s    size=%s    comment='%s'\n", val.GetId().GetId(), val.GetFilename(), created, prettifySize(val.GetSize()), val.GetComment())
	}
}
