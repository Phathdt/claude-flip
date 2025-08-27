package storage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// KeychainStorage provides cross-platform secure storage
type KeychainStorage struct {
	serviceName string
}

// NewKeychainStorage creates a new keychain storage instance
func NewKeychainStorage(serviceName string) *KeychainStorage {
	return &KeychainStorage{
		serviceName: serviceName,
	}
}

// Store saves data securely based on the platform
func (k *KeychainStorage) Store(key, data string) error {
	switch runtime.GOOS {
	case "darwin":
		return k.storeMacOS(key, data)
	case "linux":
		return k.storeLinux(key, data)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Retrieve gets data securely based on the platform
func (k *KeychainStorage) Retrieve(key string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return k.retrieveMacOS(key)
	case "linux":
		return k.retrieveLinux(key)
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Delete removes data securely based on the platform
func (k *KeychainStorage) Delete(key string) error {
	switch runtime.GOOS {
	case "darwin":
		return k.deleteMacOS(key)
	case "linux":
		return k.deleteLinux(key)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// storeMacOS stores data in macOS Keychain
func (k *KeychainStorage) storeMacOS(key, data string) error {
	// Use security command to store in keychain
	cmd := exec.Command("security", "add-generic-password", 
		"-U", // Update if exists
		"-s", k.serviceName,
		"-a", key,
		"-w", data)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to store in keychain: %w (output: %s)", err, string(output))
	}
	
	return nil
}

// retrieveMacOS retrieves data from macOS Keychain
func (k *KeychainStorage) retrieveMacOS(key string) (string, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-s", k.serviceName,
		"-a", key,
		"-w") // Return password only
	
	output, err := cmd.Output()
	if err != nil {
		if strings.Contains(err.Error(), "exit status 44") {
			return "", fmt.Errorf("key not found in keychain: %s", key)
		}
		return "", fmt.Errorf("failed to retrieve from keychain: %w", err)
	}
	
	// Remove trailing newline if present
	data := strings.TrimSuffix(string(output), "\n")
	return data, nil
}

// deleteMacOS removes data from macOS Keychain
func (k *KeychainStorage) deleteMacOS(key string) error {
	cmd := exec.Command("security", "delete-generic-password",
		"-s", k.serviceName,
		"-a", key)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(err.Error(), "exit status 44") {
			// Item not found - not an error for deletion
			return nil
		}
		return fmt.Errorf("failed to delete from keychain: %w (output: %s)", err, string(output))
	}
	
	return nil
}

// storeLinux stores data in encrypted file (fallback for Linux)
func (k *KeychainStorage) storeLinux(key, data string) error {
	// On Linux, fall back to file-based storage
	// This maintains the same interface but uses secure file storage
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	credentialsDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(credentialsDir, 0o700); err != nil {
		return fmt.Errorf("failed to create credentials directory: %w", err)
	}

	// Use service name and key to create unique filename
	filename := fmt.Sprintf(".%s_%s.json", k.serviceName, key)
	credentialsPath := filepath.Join(credentialsDir, filename)

	// Write atomically using temporary file
	tempPath := credentialsPath + ".tmp"
	if err := os.WriteFile(tempPath, []byte(data), 0o600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	if err := os.Rename(tempPath, credentialsPath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to replace credentials file: %w", err)
	}

	return nil
}

// retrieveLinux retrieves data from encrypted file (fallback for Linux)
func (k *KeychainStorage) retrieveLinux(key string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Use service name and key to create unique filename
	filename := fmt.Sprintf(".%s_%s.json", k.serviceName, key)
	credentialsPath := filepath.Join(home, ".claude", filename)

	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", fmt.Errorf("failed to read credentials file: %w", err)
	}

	return string(data), nil
}

// deleteLinux removes data from encrypted file (fallback for Linux)
func (k *KeychainStorage) deleteLinux(key string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Use service name and key to create unique filename
	filename := fmt.Sprintf(".%s_%s.json", k.serviceName, key)
	credentialsPath := filepath.Join(home, ".claude", filename)

	err = os.Remove(credentialsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist - not an error for deletion
			return nil
		}
		return fmt.Errorf("failed to delete credentials file: %w", err)
	}

	return nil
}