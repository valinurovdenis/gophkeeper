// Package encryption contains methods for encrypt file data.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"

	"github.com/valinurovdenis/gophkeeper/internal/app/config"
)

// Singleton variables for pem keys.
var (
	serverPrivateKey *[]byte = nil
	serverPublicKey  *[]byte = nil
	clientPrivateKey *[]byte = nil
	clientPublicKey  *[]byte = nil
	appConfig                = config.DefaultConfig
)

// Generate private, public rsa keys.
func getRsaKeys() ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate rsa key: %w", err)
	}

	privateKeyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyDER,
	}
	privateKeyPEM := pem.EncodeToMemory(privateKeyBlock)

	publicKey := &privateKey.PublicKey

	publicKeyDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot convert public rsa key: %w", err)
	}
	publicKeyBlock := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyDER,
	}
	publicKeyPEM := pem.EncodeToMemory(publicKeyBlock)

	return privateKeyPEM, publicKeyPEM, nil
}

// Saves key to file.
func SaveKeyToFile(key []byte, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Write(key)
	return nil
}

// Creates public,private keys pair for server|client if not exists.
func CreateKeysIfAbsent(isServer bool) error {
	var privateKeyPath string
	var publicKeyPath string
	if isServer {
		privateKeyPath = appConfig.ServerPrivateKeyPath
		publicKeyPath = appConfig.ServerPublicKeyPath
	} else {
		privateKeyPath = appConfig.ClientPrivateKeyPath
		publicKeyPath = appConfig.ClientPublicKeyPath
	}
	if _, err := os.Stat(privateKeyPath); err != nil {
		privateKey, publicKey, err := getRsaKeys()
		if err != nil {
			return err
		}
		SaveKeyToFile(privateKey, privateKeyPath)
		SaveKeyToFile(publicKey, publicKeyPath)
	}
	return nil
}

// Get pem key from file and cache it in singleton.
func getCachedKeyFromFile(filepath string, key *[]byte) func() []byte {
	return func() []byte {
		if key == nil {
			data, err := os.ReadFile(filepath)
			if err != nil {
				return nil
			}
			key = &data
		}
		return *key
	}
}

// Getters for pem keys from file.
var (
	ServerPrivateKey = func() []byte { return []byte("mock") }
	ServerPublicKey  = func() []byte { return []byte("mock") }
	ClientPrivateKey = func() []byte { return []byte("mock") }
	ClientPublicKey  = func() []byte { return []byte("mock") }
)

// Init encryption data.
func InitData() {
	appConfig = config.GetConfig()
	ServerPrivateKey = getCachedKeyFromFile(appConfig.ServerPrivateKeyPath, serverPrivateKey)
	ServerPublicKey = getCachedKeyFromFile(appConfig.ServerPublicKeyPath, serverPublicKey)
	ClientPrivateKey = getCachedKeyFromFile(appConfig.ClientPrivateKeyPath, clientPrivateKey)
	ClientPublicKey = getCachedKeyFromFile(appConfig.ClientPublicKeyPath, clientPublicKey)
}

// Encrypt symmetric file encryption key with rsa public key.
func EncryptFileEncryptionKey(fileKey, encryptionKey []byte) ([]byte, error) {
	if len(encryptionKey) == 0 {
		return nil, fmt.Errorf("empty encryption key, relogin may required")
	}
	block, _ := pem.Decode(encryptionKey)
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to get public key")
	}
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, fileKey, nil)
}

// Decrypt symmetric file encryption key with rsa private key.
func DecryptFileEncryptionKey(encryptedFileKey, encryptionKey []byte) ([]byte, error) {
	if len(encryptionKey) == 0 {
		return nil, fmt.Errorf("empty encryption key, relogin may required")
	}
	block, _ := pem.Decode(encryptionKey)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedFileKey, nil)
}

// Generate symmetric file encryption key.
func GenerateSymmetricFileEncryptionKey() ([]byte, error) {
	key := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// Encrypt file data with symmetric file encryption key.
// Preservs file data length.
func EncryptFileData(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(block, key)
	ciphertext := make([]byte, len(data))
	stream.XORKeyStream(ciphertext, data)
	return ciphertext, nil
}

// Decrypt file data with symmetric file encryption key.
// Preservs file data length.
func DecryptFileData(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, key)

	decrypted := make([]byte, len(data))
	stream.XORKeyStream(decrypted, data)
	return decrypted, nil
}
