// Package auth for authorization middlewares.
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/valinurovdenis/gophkeeper/internal/app/userstorage"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Generate password hash with salt.
func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// Compare hash and password.
func ComparePasswordHash(hashedPassword []byte, originalPassword string) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, []byte(originalPassword))
}

// Get login from incoming context.
func GetLoginFromContext(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		login := md.Get("login")
		if len(login) > 0 {
			return md.Get("login")[0]
		}
	}
	return ""
}

// Struct for parsing user id from jwt token.
type Claims struct {
	jwt.RegisteredClaims
	Login string
}

const tokenExpiration = time.Hour * 3

// Class for authentication via jwt tokens.
type JwtAuthenticator struct {
	SecretKey   string
	UserStorage userstorage.UserStorage
}

// Returns new authenticator.
// Requires secret key for jwt and storage for generatings user ids.
func NewAuthenticator(secretKey string, userStorage userstorage.UserStorage) *JwtAuthenticator {
	return &JwtAuthenticator{
		SecretKey:   secretKey,
		UserStorage: userStorage,
	}
}

// Builds jwt string from given user id.
func (a *JwtAuthenticator) BuildJWTString(login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiration)),
		},
		Login: login,
	})

	tokenString, err := token.SignedString([]byte(a.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Parses user id from jwt token string.
// Returns error if no jwt token is not valid.
func (a *JwtAuthenticator) GetLogin(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(a.SecretKey), nil
		})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	return claims.Login, nil
}

// Middleware checks whether there is authorization cookie with valid user.
// Returns 401 if valid user not found.
func (a *JwtAuthenticator) CheckAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	authorization := md.Get("Authorization")
	userAuthenticated := false
	if ok && len(authorization) > 0 {
		login, err := a.GetLogin(authorization[0])
		if err == nil {
			md = metadata.Pairs("login", login)
			ctx = metadata.NewIncomingContext(ctx, md)
			userAuthenticated = true
		}
	}
	if !userAuthenticated {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized")
	}
	return handler(ctx, req)
}

type serverStreamWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (w serverStreamWrapper) Context() context.Context { return w.ctx }

// Same as CheckAuth for streaming.
func (a *JwtAuthenticator) CheckStreamAuth(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	md, ok := metadata.FromIncomingContext(stream.Context())
	authorization := md.Get("Authorization")
	userAuthenticated := false
	var wrappedStream serverStreamWrapper
	if ok && len(authorization) > 0 {
		login, err := a.GetLogin(authorization[0])
		if err == nil {
			md = metadata.New(map[string]string{"login": login})
			ctx := metadata.NewIncomingContext(stream.Context(), md)
			wrappedStream = serverStreamWrapper{stream, ctx}
			userAuthenticated = true
		}
	}
	if !userAuthenticated {
		return status.Errorf(codes.Unauthenticated, "Unauthorized")
	}
	return handler(srv, wrappedStream)
}
