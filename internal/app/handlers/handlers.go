// Package handlers contains gophkeeper grpc handlers.
package handlers

import (
	"context"
	"slices"

	"github.com/valinurovdenis/gophkeeper/internal/app/auth"
	"github.com/valinurovdenis/gophkeeper/internal/app/logger"
	"github.com/valinurovdenis/gophkeeper/internal/app/service"
	"github.com/valinurovdenis/gophkeeper/internal/app/userstorage"
	pb "github.com/valinurovdenis/gophkeeper/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GophKeeperHandlerGrpc struct {
	pb.UnimplementedGophKeeperServiceServer
	service     service.GophKeeperService
	auth        auth.JwtAuthenticator
	userStorage userstorage.UserStorage
}

func NewGophKeeperHandler(service service.GophKeeperService, auth auth.JwtAuthenticator, userStorage userstorage.UserStorage) *GophKeeperHandlerGrpc {
	return &GophKeeperHandlerGrpc{service: service, auth: auth, userStorage: userStorage}
}

const (
	RegisterMethod = "/gophkeeper.GophKeeperService/Register"
	LoginMethod    = "/gophkeeper.GophKeeperService/Login"
)

var authMethods = []string{RegisterMethod, LoginMethod}

// Defines handlers with interceptors.
func KeeperGrpcRouter(gophKeeperHandler GophKeeperHandlerGrpc) *grpc.Server {
	authorizationInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		if !slices.Contains(authMethods, info.FullMethod) {
			return gophKeeperHandler.auth.CheckAuth(ctx, req, info, handler)
		}
		return handler(ctx, req)
	}
	authorizationStreamInterceptor := func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !slices.Contains(authMethods, info.FullMethod) {
			return gophKeeperHandler.auth.CheckStreamAuth(srv, stream, info, handler)
		}
		return handler(srv, stream)
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(logger.RequestLoggerInterceptor(), authorizationInterceptor),
		grpc.ChainStreamInterceptor(logger.RequestStreamLoggerInterceptor(), authorizationStreamInterceptor))
	pb.RegisterGophKeeperServiceServer(srv, &gophKeeperHandler)

	return srv
}

func (h *GophKeeperHandlerGrpc) Register(ctx context.Context, user *pb.UserCred) (*emptypb.Empty, error) {
	var err error
	var hashedPassword []byte
	if hashedPassword, err = auth.HashPassword(user.Password); err == nil {
		err = h.userStorage.AddUser(ctx, userstorage.User{Login: user.Login, PasswordHash: hashedPassword})
	}
	if err != nil {
		if err == userstorage.ErrConflictUserLogin {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	token, err := h.auth.BuildJWTString(user.Login)
	if err != nil {
		return nil, err
	}
	grpc.SetHeader(ctx, metadata.Pairs("Authorization", token))
	return nil, nil
}

func (h *GophKeeperHandlerGrpc) Login(ctx context.Context, user *pb.UserCred) (*emptypb.Empty, error) {
	existingUser, err := h.userStorage.GetUser(ctx, user.Login)
	if err != nil || existingUser == nil {
		return nil, status.Errorf(codes.NotFound, "error getting user by login")
	}
	if auth.ComparePasswordHash(existingUser.PasswordHash, user.Password) != nil {
		return nil, status.Errorf(codes.PermissionDenied, "wrong password")
	}
	token, err := h.auth.BuildJWTString(user.Login)
	if err != nil {
		return nil, err
	}
	grpc.SetHeader(ctx, metadata.Pairs("Authorization", token))
	return nil, nil
}

func (h *GophKeeperHandlerGrpc) GetUserFiles(ctx context.Context, _ *emptypb.Empty) (*pb.ListFiles, error) {
	login := auth.GetLoginFromContext(ctx)
	files, err := h.service.GetUserFiles(ctx, login)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return files, nil
}

func (h *GophKeeperHandlerGrpc) UploadFile(srv pb.GophKeeperService_UploadFileServer) error {
	login := auth.GetLoginFromContext(srv.Context())
	err := h.service.UploadFile(srv, login)
	if err != nil {
		return status.Errorf(codes.Internal, err.Error())
	}
	return nil
}

func (h *GophKeeperHandlerGrpc) DownloadFile(fileId *pb.FileId, srv pb.GophKeeperService_DownloadFileServer) error {
	login := auth.GetLoginFromContext(srv.Context())
	err := h.service.DownloadFile(fileId, srv, login)
	if err != nil {
		if err == service.ErrNotOwn {
			return status.Errorf(codes.PermissionDenied, err.Error())
		}
		return status.Errorf(codes.Internal, err.Error())
	}
	return nil
}

func (h *GophKeeperHandlerGrpc) DeleteFile(ctx context.Context, fileId *pb.FileId) (*emptypb.Empty, error) {
	login := auth.GetLoginFromContext(ctx)
	err := h.service.DeleteFile(ctx, fileId, login)
	if err != nil {
		if err == service.ErrNotOwn {
			return nil, status.Errorf(codes.PermissionDenied, err.Error())
		}
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return nil, nil
}
