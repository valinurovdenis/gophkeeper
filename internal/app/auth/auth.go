// Package auth for authorization middlewares.
package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
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
func GetVarFromContext(ctx context.Context, name string) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		login := md.Get(name)
		if len(login) > 0 {
			return md.Get(name)[0]
		}
	}
	return ""
}

// Struct for parsing user id from jwt token.
type Claims struct {
	jwt.RegisteredClaims
	Login     string
	PublicKey string
}

const tokenExpiration = time.Hour * 3

// Class for authentication via jwt tokens.
type JwtAuthenticator struct {
	SecretKey string
}

// Returns new authenticator.
// Requires secret key for jwt and storage for generatings user ids.
func NewAuthenticator(secretKey string) *JwtAuthenticator {
	return &JwtAuthenticator{
		SecretKey: secretKey,
	}
}

// Builds jwt string from given user id.
func (a *JwtAuthenticator) BuildJWTString(login string, publicKey []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiration)),
		},
		Login:     login,
		PublicKey: string(publicKey),
	})

	tokenString, err := token.SignedString([]byte(a.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Parses user id from jwt token string.
// Returns error if no jwt token is not valid.
func (a *JwtAuthenticator) GetJwtClaims(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(a.SecretKey), nil
		})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (a *JwtAuthenticator) getAuthContext(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	authorization := md.Get("Authorization")
	if ok && len(authorization) > 0 {
		claims, err := a.GetJwtClaims(authorization[0])
		if err == nil {
			md = metadata.Pairs("login", claims.Login, "public_key", claims.PublicKey)
			return metadata.NewIncomingContext(ctx, md), nil
		}
	}
	return ctx, fmt.Errorf("unauthorized")
}

// Middleware checks whether there is authorization cookie with valid user.
// Returns 401 if valid user not found.
func (a *JwtAuthenticator) CheckAuth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {

	authCtx, err := a.getAuthContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Unauthorized")
	}
	return handler(authCtx, req)
}

type serverStreamWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (w serverStreamWrapper) Context() context.Context { return w.ctx }

// Same as CheckAuth for streaming.
func (a *JwtAuthenticator) CheckStreamAuth(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	authCtx, err := a.getAuthContext(stream.Context())
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "Unauthorized")
	}
	return handler(srv, serverStreamWrapper{stream, authCtx})
}
