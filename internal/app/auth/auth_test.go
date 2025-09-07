package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJwtAuthenticator_testJWTToken(t *testing.T) {
	login := "kulebaka"
	publicKey := "qwerty"
	secretKey := "asdf"
	authenticator := NewAuthenticator(secretKey)
	validTokenString, err := authenticator.BuildJWTString(login, []byte(publicKey))
	require.NoError(t, err)
	wrongSigningMethodTokenString, _ := jwt.NewWithClaims(jwt.SigningMethodRS256,
		jwt.RegisteredClaims{}).SignedString([]byte(secretKey))

	tests := []struct {
		name        string
		tokenString string
		wantErr     bool
	}{
		{name: "valid", tokenString: validTokenString, wantErr: false},
		{name: "invalid", tokenString: "asdf", wantErr: true},
		{name: "wrongMethod", tokenString: wrongSigningMethodTokenString, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := authenticator.GetJwtClaims(tt.tokenString)
			require.Equal(t, err != nil, tt.wantErr)
			if !tt.wantErr {
				assert.Equal(t, login, res.Login)
				assert.Equal(t, publicKey, res.PublicKey)
			}
		})
	}
}
