// Package userstorage for storing and generating user ids.
package userstorage

import "context"

type User struct {
	Login        string
	PasswordHash []byte
}

// Storage can generate uuid for new user with no collision.
//
//go:generate mockery --name UserStorage
type UserStorage interface {
	// Method for adding new user.
	AddUser(ctx context.Context, user User) error

	// Method for adding new user.
	GetUser(ctx context.Context, login string) (*User, error)
}
