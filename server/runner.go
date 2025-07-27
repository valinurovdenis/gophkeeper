// Package runner for running service with given config.
package main

import (
	"database/sql"
	"net"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/valinurovdenis/gophkeeper/internal/app/auth"
	"github.com/valinurovdenis/gophkeeper/internal/app/config"
	"github.com/valinurovdenis/gophkeeper/internal/app/filestorage"
	"github.com/valinurovdenis/gophkeeper/internal/app/handlers"
	"github.com/valinurovdenis/gophkeeper/internal/app/logger"
	"github.com/valinurovdenis/gophkeeper/internal/app/metadatastorage"
	"github.com/valinurovdenis/gophkeeper/internal/app/service"
	"github.com/valinurovdenis/gophkeeper/internal/app/userstorage"
)

// Initialize db connection.
func GetDB() *sql.DB {
	config := config.GetConfig()
	db, err := sql.Open("pgx", config.Database)
	if err != nil {
		panic(err)
	}
	return db
}

// Initialize s3 connection.
func GetS3() *filestorage.S3FileStorage {
	s3Storage, err := filestorage.NewS3FileStorage()
	if err != nil {
		panic(err)
	}
	return s3Storage
}

// Runs keeper service with given config.
func Run() error {
	config := config.GetConfig()

	if err := logger.Initialize(config.LogLevel); err != nil {
		return err
	}

	db := GetDB()
	defer db.Close()
	s3Storage := GetS3()

	metaDataStorage := metadatastorage.NewPostgresqlStorageStorage(db)
	service, err := service.NewGophKeeperService(s3Storage, metaDataStorage)
	if err != nil {
		return err
	}
	userStorage := userstorage.NewPostgresqlUserStorage(db)
	auth := auth.NewAuthenticator(config.SecretKey, userStorage)
	grpcHandler, err := handlers.NewGophKeeperHandler(*service, *auth, userStorage)
	if err != nil {
		return err
	}
	grpcSrv := handlers.KeeperGrpcRouter(*grpcHandler)
	listen, _ := net.Listen("tcp", config.LocalURL)
	return grpcSrv.Serve(listen)
}
