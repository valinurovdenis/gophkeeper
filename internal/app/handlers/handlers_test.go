package handlers

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valinurovdenis/gophkeeper/internal/app/auth"
	"github.com/valinurovdenis/gophkeeper/internal/app/service"
	"github.com/valinurovdenis/gophkeeper/internal/mocks"
	pb "github.com/valinurovdenis/gophkeeper/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const bufSize = 1024 * 1024
const secretKey = "secret_key"

func initHandlers(mockMetadataStorage *mocks.MetadataStorage,
	mockStreamingFileStorage *mocks.StreamingFileStorage,
	mockUserStorage *mocks.UserStorage,
	auth *auth.JwtAuthenticator) (*grpc.Server, *bufconn.Listener) {

	lis := bufconn.Listen(bufSize)
	service, _ := service.NewGophKeeperService(mockStreamingFileStorage, mockMetadataStorage)
	grpcHandler := GophKeeperHandlerGrpc{service: *service, auth: *auth, userStorage: mockUserStorage}
	grpcSrv := KeeperGrpcRouter(grpcHandler)
	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
	return grpcSrv, lis
}

func getGrpcConn(t *testing.T, lis *bufconn.Listener) *grpc.ClientConn {
	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	return conn
}

func getStatusFromGrpcError(t *testing.T, err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	if e, ok := status.FromError(err); ok {
		return e.Code()
	} else {
		t.Fatalf("Не получилось распарсить ошибку %v", err)
		return codes.Internal
	}
}

func TestShortenerHandlerGrpc_TestUnauthorized(t *testing.T) {
	mockMetadataStorage := mocks.NewMetadataStorage(t)
	mockStreamingFileStorage := mocks.NewStreamingFileStorage(t)
	mockUserStorage := mocks.NewUserStorage(t)
	auth := auth.NewAuthenticator(secretKey)
	grpcSrv, lis := initHandlers(mockMetadataStorage, mockStreamingFileStorage, mockUserStorage, auth)
	conn := getGrpcConn(t, lis)
	defer conn.Close()
	grpcClient := pb.NewGophKeeperServiceClient(conn)
	_, err := grpcClient.GetUserFiles(context.Background(), &emptypb.Empty{})
	require.NotNil(t, err)
	require.Equal(t, codes.Unauthenticated, getStatusFromGrpcError(t, err))
	grpcSrv.Stop()
}
