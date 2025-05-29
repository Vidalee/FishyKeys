package shamir

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"github.com/hashicorp/vault/shamir"
)

// GenerateSecret generates a random AES-256 secret
func GenerateSecret() ([]byte, error) {
	secret := make([]byte, 32) // AES-256
	_, err := rand.Read(secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// SplitSecret splits a secret into n shares, requiring k shares to reconstruct
func SplitSecret(secret []byte, n, k int) ([][]byte, error) {
	if len(secret) == 0 {
		return nil, errors.New("secret is empty")
	}
	return shamir.Split(secret, n, k)
}

// CombineShares combines k shares to reconstruct the original secret
func CombineShares(shares [][]byte) ([]byte, error) {
	return shamir.Combine(shares)
}

// Encrypt encrypts data with AES-GCM using the provided key
func Encrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts AES-GCM encrypted data with the provided key
func Decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}
