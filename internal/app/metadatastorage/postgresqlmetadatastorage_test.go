package metadatastorage

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/valinurovdenis/gophkeeper/internal/proto"
)

func TestPostgresqlStorage_GetFileById(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	created := time.Now()
	fileId := pb.FileId{Id: "id"}
	key := []byte("key")
	fileInfo := pb.FileInfo{Id: &fileId, Login: "login", Filename: "name", Comment: "comment", Created: uint64(created.Unix()), Size: 1, EncryptionKey: key}

	storage := NewPostgresqlStorageStorage(db)
	tests := []struct {
		name    string
		wantErr bool
		isFound bool
	}{
		{name: "get_error", wantErr: true, isFound: false},
		{name: "get_not_empty", wantErr: false, isFound: true},
		{name: "get_empty", wantErr: false, isFound: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mock.ExpectQuery("SELECT").WillReturnError(&pgconn.PgError{})
			} else if tt.isFound {
				mock.ExpectQuery("SELECT").WillReturnRows(
					sqlmock.NewRows([]string{"id", "login", "filename", "comment", "created", "size", "enctyprion_key"}).AddRow(
						"id", "login", "name", "comment", created, 1, key))
			} else {
				mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))
			}
			got, err := storage.GetFileById(context.Background(), "id")
			if !tt.wantErr && tt.isFound {
				require.NoError(t, err)
				assert.Equal(t, &fileInfo, got)
			} else {
				assert.NotEqual(t, err, nil)
			}
		})
	}
}

func TestPostgresqlStorage_GetFilesByLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	created := time.Now()

	key := []byte("key")
	listFiles := pb.ListFiles{}
	for _, id := range []string{"id1", "id2"} {
		fileId := pb.FileId{Id: id}
		fileInfo := pb.FileInfo{Id: &fileId, Login: "login", Filename: "name", Comment: "comment", Created: uint64(created.Unix()), Size: 1, EncryptionKey: key}
		listFiles.Files = append(listFiles.Files, &fileInfo)
	}

	storage := NewPostgresqlStorageStorage(db)
	tests := []struct {
		name    string
		wantErr bool
		isFound bool
	}{
		{name: "get_error", wantErr: true, isFound: false},
		{name: "get_not_empty", wantErr: false, isFound: true},
		{name: "get_empty", wantErr: false, isFound: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mock.ExpectQuery("SELECT").WillReturnError(&pgconn.PgError{})
			} else if tt.isFound {
				mock.ExpectQuery("SELECT").WillReturnRows(
					sqlmock.NewRows([]string{"id", "login", "filename", "comment", "created", "size", "enctyprion_key"}).AddRows(
						[]driver.Value{"id1", "login", "name", "comment", created, 1, key}, []driver.Value{"id2", "login", "name", "comment", created, 1, key}))
			} else {
				mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))
			}
			got, err := storage.GetFilesByLogin(context.Background(), "login")
			if !tt.wantErr {
				require.NoError(t, err)
				if tt.isFound {
					assert.Equal(t, &listFiles, got)
				} else {
					assert.Empty(t, got)
				}
			} else {
				assert.NotEqual(t, err, nil)
			}
		})
	}
}

func TestPostgresqlStorage_AddFileInfo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	created := time.Now()
	fileId := pb.FileId{Id: "id"}
	key := []byte("key")
	fileInfo := pb.FileInfo{Id: &fileId, Login: "login", Filename: "name", Comment: "comment", Created: uint64(created.Unix()), Size: 1, EncryptionKey: key}

	storage := NewPostgresqlStorageStorage(db)
	tests := []struct {
		name       string
		wantErr    bool
		alreadyHas bool
	}{
		{name: "add_error", wantErr: true, alreadyHas: false},
		{name: "add_already_exists", wantErr: false, alreadyHas: true},
		{name: "add_good", wantErr: false, alreadyHas: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mock.ExpectExec("INSERT into fileinfo").WillReturnError(&pgconn.PgError{Code: pgerrcode.ConnectionException})
			} else if tt.alreadyHas {
				mock.ExpectExec("INSERT into fileinfo").WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
			} else {
				mock.ExpectExec("INSERT into fileinfo").WillReturnResult(sqlmock.NewResult(1, 1))
			}
			err := storage.AddFileInfo(context.Background(), &fileInfo)
			if tt.wantErr {
				assert.NotEqual(t, err, nil)
			} else if tt.alreadyHas {
				assert.ErrorIs(t, err, ErrConflictMetaId)
			} else {
				assert.Equal(t, err, nil)
			}
		})
	}
}

func TestPostgresqlStorage_DeleteFileInfo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	storage := NewPostgresqlStorageStorage(db)
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "delete_error", wantErr: true},
		{name: "delete_good", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				mock.ExpectExec("DELETE").WillReturnError(&pgconn.PgError{})
			} else {
				mock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(1, 1))
			}
			err := storage.DeleteFileInfo(context.Background(), "id")
			if tt.wantErr {
				assert.NotEqual(t, err, nil)
			} else {
				assert.Equal(t, err, nil)
			}
		})
	}
}

func TestPostgresqlStorage_Ping(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	storage := NewPostgresqlStorageStorage(db)
	err = storage.Ping()
	assert.Equal(t, err, nil)
}

func TestPostgresqlStorage_init(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE fileinfo").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE INDEX login_index").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	storage := NewPostgresqlStorageStorage(db)
	storage.init()
}
