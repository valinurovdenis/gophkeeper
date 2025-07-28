package metadatastorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	pb "github.com/valinurovdenis/gophkeeper/internal/proto"
)

// Error in case file with given id already has been saved.
var ErrConflictMetaId = errors.New("conflicting id")

type PostgresqlStorage struct {
	DB *sql.DB
}

func NewPostgresqlStorageStorage(db *sql.DB) *PostgresqlStorage {
	ret := &PostgresqlStorage{DB: db}
	ret.init()
	return ret
}

func (s *PostgresqlStorage) init() error {
	tx, err := s.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tx.Exec(`CREATE TABLE fileinfo("id" TEXT PRIMARY KEY, "login" TEXT NOT NULL CHECK ("login" <> ''), "filename" TEXT, "comment" TEXT, "created" TIMESTAMP, "modified" TIMESTAMP, "size" INT, "encryption_key" bytea)`)
	tx.Exec(`CREATE INDEX login_index ON fileinfo USING btree(login)`)
	return tx.Commit()
}

func (s *PostgresqlStorage) GetFileById(ctx context.Context, fileId string) (*pb.FileInfo, error) {
	row := s.DB.QueryRowContext(ctx,
		"SELECT id, login, filename, comment, created, size, encryption_key FROM fileinfo WHERE id = $1", fileId)
	file := pb.FileInfo{}
	var created time.Time
	var id string
	err := row.Scan(&id, &file.Login, &file.Filename, &file.Comment, &created, &file.Size, &file.EncryptionKey)
	file.Id = &pb.FileId{Id: id}
	file.Created = uint64(created.Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to scan rows: %w", err)
	}
	return &file, nil
}

func (s *PostgresqlStorage) GetFilesByLogin(ctx context.Context, login string) (*pb.ListFiles, error) {
	var files []*pb.FileInfo
	rows, err := s.DB.QueryContext(ctx,
		"SELECT id, login, filename, comment, created, size, encryption_key FROM fileinfo WHERE login = $1", login)
	if err != nil {
		return nil, fmt.Errorf("failed to begin select query: %w", err)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to get rows: %w", rows.Err())
	}
	defer rows.Close()
	for rows.Next() {
		file := pb.FileInfo{}
		var created time.Time
		var id string
		err = rows.Scan(&id, &file.Login, &file.Filename, &file.Comment, &created, &file.Size, &file.EncryptionKey)
		file.Id = &pb.FileId{Id: id}
		file.Created = uint64(created.Unix())
		if err != nil {
			return nil, err
		}
		files = append(files, &file)
	}

	return &pb.ListFiles{Files: files}, nil
}

func (s *PostgresqlStorage) AddFileInfo(ctx context.Context, fileInfo *pb.FileInfo) error {
	_, err := s.DB.ExecContext(ctx,
		"INSERT into fileinfo (id, login, filename, comment, created, size, encryption_key) VALUES($1, $2, $3, $4, $5, $6, $7)",
		fileInfo.GetId().GetId(), fileInfo.GetLogin(), fileInfo.GetFilename(), fileInfo.GetComment(),
		time.Unix(int64(fileInfo.GetCreated()), 0), fileInfo.GetSize(), fileInfo.GetEncryptionKey())
	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.UniqueViolation {
		err = ErrConflictMetaId
	}
	return err
}

func (s *PostgresqlStorage) DeleteFileInfo(ctx context.Context, fileId string) error {
	query := `DELETE from fileinfo where id = $1`
	_, err := s.DB.ExecContext(ctx, query, fileId)

	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresqlStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return s.DB.PingContext(ctx)
}
