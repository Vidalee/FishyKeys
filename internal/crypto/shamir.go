package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
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

// EncryptWithKey encrypts data with AES-GCM using the provided key
func EncryptWithKey(key, plaintext []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	cipherTextEncoded := base64.StdEncoding.EncodeToString(ciphertext)
	return cipherTextEncoded, nil
}

// Encrypt encrypts data with AES-GCM using the unlocked KeyManager
func Encrypt(keyManager *KeyManager, plaintext []byte) (string, error) {
	if keyManager == nil {
		return "", errors.New("key manager is nil")
	}

	key, err := keyManager.GetMasterKey()
	if err != nil {
		return "", err
	}

	return EncryptWithKey(key, plaintext)
}

// DecryptWithKey decrypts AES-GCM encrypted data with the provided key
func DecryptWithKey(key []byte, encodedCiphertext string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, err
	}
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

// Decrypt decrypts AES-GCM encrypted data with the unlocked KeyManager
func Decrypt(keyManager *KeyManager, encodedCiphertext string) ([]byte, error) {
	if keyManager == nil {
		return nil, errors.New("key manager is nil")
	}

	key, err := keyManager.GetMasterKey()
	if err != nil {
		return nil, err
	}

	return DecryptWithKey(key, encodedCiphertext)
}
