// Package service contains methods for file processing.
package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/valinurovdenis/gophkeeper/internal/app/encryption"
	"github.com/valinurovdenis/gophkeeper/internal/mocks"
	pb "github.com/valinurovdenis/gophkeeper/internal/proto"
	"google.golang.org/grpc"
)

type serverStreamMock struct {
	grpc.ServerStream
	ctx      context.Context
	t        *testing.T
	fileInfo *pb.FileInfo
	file     *[]byte
}

func (s serverStreamMock) Recv() (*pb.FileStream, error) {
	return &pb.FileStream{Data: &pb.FileStream_Info{Info: s.fileInfo}}, nil
}

func (s serverStreamMock) Send(filePart *pb.FileStream) error {
	if filePart.GetInfo() != nil {
		require.Equal(s.t, &pb.FileInfo{}, s.fileInfo, "file info should be send exactly once")
		*s.fileInfo = *filePart.GetInfo()
		return nil
	}
	require.NotEqual(s.t, filePart.GetChunkData(), nil, "file stream should contain chunk data or file info")
	*s.file = append(*s.file, filePart.GetChunkData()...)
	return nil
}

func (s serverStreamMock) SendAndClose(resp *pb.UploadResponse) error {
	require.NotNil(s.t, resp.Id)
	s.fileInfo.Id = resp.Id
	return nil
}

func (s serverStreamMock) Context() context.Context {
	return s.ctx
}

const ServerPublicKey = `-----BEGIN RSA PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC7n40jGDHksr1YzfDhoIxIgOZU
pdfZVVNGdzxkv86Lwjw4oWmI4pQed5E3j+bbk5qxiaFawL55zveVXIQzDaUoUr9s
UGJ1b1petl6vwKS/n8PjbU/o6aZcNlyt8VA7QWAWgeXYgg3meoSyj0wyL10L73iw
4l3O507MDsd2fYprMQIDAQAB
-----END RSA PUBLIC KEY-----`
const ServerPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQC7n40jGDHksr1YzfDhoIxIgOZUpdfZVVNGdzxkv86Lwjw4oWmI
4pQed5E3j+bbk5qxiaFawL55zveVXIQzDaUoUr9sUGJ1b1petl6vwKS/n8PjbU/o
6aZcNlyt8VA7QWAWgeXYgg3meoSyj0wyL10L73iw4l3O507MDsd2fYprMQIDAQAB
AoGAX9n2F6y/qI+r7hdf7VTA9jVr9mi3ah+OKJy3rNzUn0++xkuoB7eBZkM9W/5X
OWwiBntChIOdi8sxbwvRuedJrPzC1b8t26GZ0IeeJb+iIUIf+421pl0OTdWrYGGP
dZnHLQgrBpnR960TgqRhNOuxGHOsoZTuloQBy6q5AU5cQkUCQQDP3HjWdaKVk+vm
0AEO+ArN2n1LExkLdEcSekO/3DacRBWlyf8AaOEufa56TcF6U70smWBPs4RmnO+O
EK1+0O0HAkEA5xM5j4FnllklZefXUdDpXLmuk1RyEz1HADLGdaCWEJmxHZ7QdF/N
3LqcP5lq1fH2gultuSmC8xQWA1Ph24CQBwJBAJIQhqWFcmOT17CROD0xlj4DrAnm
eLHw2sSkQBmBgKqcuW2QHW5HRP2recEeBLiWQZgmi2RWbNLCsx/snk5AOF8CQQDQ
fQUTWOOYwOhAUPVyqXbUpehAoBGpEEHOiQGNQg4D/lfS7OcSCRraDDlMHOVLEdyk
c27/gNfY8IeICxgej5njAkAXxrn9zfSXUGlYQ/NwkEKkcczFujz27xSNXsmSHdIo
Dfmw4IqJBkNzSbEDbYaswoMwxaj1FODIl081XarYIuZb
-----END RSA PRIVATE KEY-----`
const ClientPublicKey = `-----BEGIN RSA PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDaXWLozpyca9XiawOvxgcx54HT
90z1UwpC+wpt83AaS1/Y2GsrXdL0ZclaoLGKlzoyR1oEzHxLVJOa7Kg1o1he0vjO
nOM2w0tfMfSJSCNL/bFGTpVn77sxGi1I5sskgygswVuLqBWvUSNyRBYRKBmpMgDg
+BEQZf07Rm2v4AyGjQIDAQAB
-----END RSA PUBLIC KEY-----`
const ClientPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDaXWLozpyca9XiawOvxgcx54HT90z1UwpC+wpt83AaS1/Y2Gsr
XdL0ZclaoLGKlzoyR1oEzHxLVJOa7Kg1o1he0vjOnOM2w0tfMfSJSCNL/bFGTpVn
77sxGi1I5sskgygswVuLqBWvUSNyRBYRKBmpMgDg+BEQZf07Rm2v4AyGjQIDAQAB
AoGBAMKKnnsQz8Af5l6gvpkT0Qcp3KWOVlbd02+XHlSUpFQYwNx8+wWYwj+Qi1Id
he8WCfgPT2ilETs/r30/yCB5VVkH9yblGNdPLoIQOR4mhH2JclPgAaQiC8T3RKuZ
24CLITdq60Aphu6fLvcBlb5Idnh4LFCmUpLmbWP2U9CvL5/BAkEA4zGJZPHAf3/R
OvhsVfUn2RG8OoBjsbPPGTGn6pxLZEVG17xQT5uHiuZeurqO9iprg2vzaUM4DTBL
TQpnqB7SHQJBAPYNR2ajwwU7o8mtuGSDVR99T/Ahicd2bPKOFO1v9g0rYYDa3HX3
l4pEyJN6YpPGzXqT47zxDC5N6ChkGzluWzECQQDAz7Kd07mduxkTpe8zSBqoUy/e
qkVxc3soE4dBSaGGGHEV+ABkf0cZ74anjFp1ueyCnWP3io+QSdMuL81m1blVAkBa
/swyJEwiak0HcAyqd3uKmsBucSjQMHbYOT16FhbsBegYTFiN9BQCGbAIApHkTvh8
5aaqoIa9tSgvj94Vnj9xAkAJ93cjW2KZKmzu/+W4VtVIoZqi9ibIBIJojx8VsbSO
U4IdnnWJ7VfzThlAkwYWp8F1s7WFcsUSGQb+f1wYqQ6r
-----END RSA PRIVATE KEY-----`

func setMockEncryption() {
	encryption.ServerPrivateKey = func() []byte {
		return []byte(ServerPrivateKey)
	}
	encryption.ServerPublicKey = func() []byte {
		return []byte(ServerPublicKey)
	}
	encryption.ClientPrivateKey = func() []byte {
		return []byte(ClientPrivateKey)
	}
	encryption.ClientPublicKey = func() []byte {
		return []byte(ClientPublicKey)
	}
}

func TestGophKeeperService_UploadFile(t *testing.T) {
	setMockEncryption()
	mockMetadataStorage := mocks.NewMetadataStorage(t)
	mockStreamingFileStorage := mocks.NewStreamingFileStorage(t)

	service, err := NewGophKeeperService(mockStreamingFileStorage, mockMetadataStorage)
	require.NoError(t, err)
	encryptionKey, _ := encryption.EncryptFileEncryptionKey([]byte("encrypt"), encryption.ServerPublicKey())
	fileInfo := pb.FileInfo{Filename: "asdf", EncryptionKey: encryptionKey, Size: 1}
	stream := serverStreamMock{ctx: context.Background(), t: t, fileInfo: &fileInfo}
	login := "kulebaka"

	mockStreamingFileStorage.On("Upload", stream, int64(1), mock.Anything).Return(nil).Once()
	mockMetadataStorage.On("AddFileInfo", stream.Context(), &fileInfo).Return(nil).Once()

	err = service.UploadFile(stream, login)
	require.NoError(t, err)
}

func TestGophKeeperService_DownloadFile(t *testing.T) {
	setMockEncryption()
	mockMetadataStorage := mocks.NewMetadataStorage(t)
	mockStreamingFileStorage := mocks.NewStreamingFileStorage(t)

	service, err := NewGophKeeperService(mockStreamingFileStorage, mockMetadataStorage)
	require.NoError(t, err)

	key := []byte("encrypt")

	login := "kulebaka"
	filename := "asdf"
	fileId := pb.FileId{Id: "12345"}
	fileBytes := make([]byte, 0)
	stream := serverStreamMock{ctx: context.Background(), t: t, fileInfo: &pb.FileInfo{}, file: &fileBytes}

	mockStreamingFileStorage.On("Download", stream, fileId.GetId()).Return(nil).Once()
	encryptionKey, _ := encryption.EncryptFileEncryptionKey(key, encryption.ServerPublicKey())
	fileInfo := pb.FileInfo{Filename: "asdf", EncryptionKey: encryptionKey, Size: 1, Login: login, Id: &fileId}
	mockMetadataStorage.On("GetFileById", stream.Context(), fileId.GetId()).Return(&fileInfo, nil).Once()

	err = service.DownloadFile(&fileId, stream, login, encryption.ClientPublicKey())
	require.NoError(t, err)
	decryptedKey, _ := encryption.DecryptFileEncryptionKey(stream.fileInfo.EncryptionKey, encryption.ClientPrivateKey())
	require.Equal(t, key, decryptedKey)
	require.Equal(t, filename, stream.fileInfo.Filename)
	require.Equal(t, login, stream.fileInfo.Login)
	require.Equal(t, &fileId, stream.fileInfo.Id)
}
