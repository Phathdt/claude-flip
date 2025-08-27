package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// encrypt encrypts data using AES-GCM with a key derived from system information
func encrypt(data []byte) ([]byte, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM
func decrypt(data []byte) ([]byte, error) {
	key, err := getEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// getEncryptionKey derives an encryption key from system-specific information
func getEncryptionKey() ([]byte, error) {
	// Use a combination of user home directory and hostname for key derivation
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	// Create a deterministic key based on system information
	keyMaterial := fmt.Sprintf("claude-flip:%s:%s", home, hostname)

	// Check if we have a stored salt, if not create one
	salt, err := getOrCreateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to get salt: %w", err)
	}

	// Use SHA256 to create a 32-byte key
	h := sha256.New()
	h.Write([]byte(keyMaterial))
	h.Write(salt)
	return h.Sum(nil), nil
}

// getOrCreateSalt gets an existing salt or creates a new one
func getOrCreateSalt() ([]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dataDir := filepath.Join(home, ".claude-flip")
	saltPath := filepath.Join(dataDir, ".salt")

	// Try to read existing salt
	if salt, err := os.ReadFile(saltPath); err == nil {
		return salt, nil
	}

	// Create new salt
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Save salt with secure permissions
	if err := os.WriteFile(saltPath, salt, 0o600); err != nil {
		return nil, fmt.Errorf("failed to save salt: %w", err)
	}

	return salt, nil
}
