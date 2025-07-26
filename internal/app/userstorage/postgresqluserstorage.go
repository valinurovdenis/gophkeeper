package userstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

// Error in case user already has been saved.
var ErrConflictUserLogin = errors.New("conflicting login")

// Generates new user id by autoincrement in postgresql.
type PostgresqlUserStorage struct {
	DB *sql.DB
}

// New postgresql user storage.
func NewPostgresqlUserStorage(db *sql.DB) *PostgresqlUserStorage {
	ret := &PostgresqlUserStorage{DB: db}
	ret.init()
	return ret
}

// Create all tables if needed.
func (s *PostgresqlUserStorage) init() error {
	tx, err := s.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	tx.Exec(`CREATE TABLE userinfo("login" TEXT PRIMARY KEY, "password_hash" BYTEA)`)
	return tx.Commit()
}

// Add new user.
func (s *PostgresqlUserStorage) AddUser(ctx context.Context, user User) error {
	_, err := s.DB.ExecContext(ctx,
		"INSERT into userinfo (login, password_hash) VALUES($1, $2)", user.Login, user.PasswordHash)
	if e, ok := err.(*pgconn.PgError); ok && e.Code == pgerrcode.UniqueViolation {
		err = ErrConflictUserLogin
	}
	return err
}

// Get user.
func (s *PostgresqlUserStorage) GetUser(ctx context.Context, login string) (*User, error) {
	row := s.DB.QueryRowContext(ctx,
		"SELECT login, password_hash FROM userinfo WHERE login = $1", login)
	var user User
	err := row.Scan(&user.Login, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}
