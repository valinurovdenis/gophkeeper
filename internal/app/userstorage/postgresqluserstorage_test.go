package userstorage

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresqlUserStorage_init(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE userinfo").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	storage := NewPostgresqlUserStorage(db)
	storage.init()
}

func TestPostgresqlUserStorage_AddUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	userInfo := User{Login: "asdf", PasswordHash: []byte("qwer")}
	storage := NewPostgresqlUserStorage(db)
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
				mock.ExpectExec("INSERT into userinfo").WillReturnError(&pgconn.PgError{Code: pgerrcode.ConnectionException})
			} else if tt.alreadyHas {
				mock.ExpectExec("INSERT into userinfo").WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
			} else {
				mock.ExpectExec("INSERT into userinfo").WillReturnResult(sqlmock.NewResult(1, 1))
			}
			err := storage.AddUser(context.Background(), userInfo)
			if tt.wantErr {
				assert.NotEqual(t, err, nil)
			} else if tt.alreadyHas {
				assert.ErrorIs(t, err, ErrConflictUserLogin)
			} else {
				assert.Equal(t, err, nil)
			}
		})
	}
}

func TestPostgresqlUserStorage_GetUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	userInfo := User{Login: "asdf", PasswordHash: []byte("qwer")}
	storage := NewPostgresqlUserStorage(db)
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
					sqlmock.NewRows([]string{"login", "password_hash"}).AddRow(
						userInfo.Login, userInfo.PasswordHash))
			} else {
				mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{}))
			}
			got, err := storage.GetUser(context.Background(), userInfo.Login)
			if !tt.wantErr && tt.isFound {
				require.NoError(t, err)
				assert.Equal(t, &userInfo, got)
			} else {
				assert.NotEqual(t, err, nil)
			}
		})
	}
}
