package ssh

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

// EncryptCredential encrypts a plaintext credential using AES-256-GCM
// with a key derived from the master password via Argon2id.
// Returns a base64-encoded string: salt(16) + nonce(12) + ciphertext.
func EncryptCredential(plaintext, masterPassword string) (string, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(masterPassword), salt, 1, 64*1024, 4, 32)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aead.Seal(nil, nonce, []byte(plaintext), nil)
	combined := append(salt, nonce...)
	combined = append(combined, ciphertext...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// DecryptCredential decrypts a credential previously encrypted with EncryptCredential.
// Returns an error if the master password is wrong or data is corrupted.
func DecryptCredential(encoded, masterPassword string) (string, error) {
	combined, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	if len(combined) < 16+12+1 {
		return "", errors.New("invalid encrypted data")
	}
	salt := combined[:16]
	nonce := combined[16:28]
	ciphertext := combined[28:]
	key := argon2.IDKey([]byte(masterPassword), salt, 1, 64*1024, 4, 32)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("decryption failed: wrong master password or corrupted data")
	}
	return string(plaintext), nil
}

// DeriveMasterKey derives a deterministic base64 key from a password using SHA-256.
// Useful for storing a verification hash without exposing the original password.
func DeriveMasterKey(password string) string {
	hash := sha256.Sum256([]byte(password))
	return base64.StdEncoding.EncodeToString(hash[:])
}
