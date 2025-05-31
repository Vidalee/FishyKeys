package crypto

import (
	"bytes"
	"testing"
)

func TestEndToEndSecretSharing(t *testing.T) {
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}

	shares, err := SplitSecret(secret, 5, 3)
	if err != nil {
		t.Fatalf("SplitSecret failed: %v", err)
	}

	recovered, err := CombineShares(shares[:3])
	if err != nil {
		t.Fatalf("CombineShares failed: %v", err)
	}

	if !bytes.Equal(secret, recovered) {
		t.Fatal("Recovered secret does not match original")
	}

	plaintext := []byte("This is a top secret message.")

	ciphertext, err := Encrypt(secret, plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := Decrypt(secret, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Fatalf("Decrypted text does not match original")
	}
}

func TestEndToEndWithWrongKey(t *testing.T) {
	secret, _ := GenerateSecret()
	wrongKey := make([]byte, 32)
	plaintext := []byte("Sensitive data")

	ciphertext, err := Encrypt(secret, plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = Decrypt(wrongKey, ciphertext)
	if err == nil {
		t.Fatal("Expected decryption error with wrong key, got nil")
	}
}
