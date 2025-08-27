package storage

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// SecureStorage defines the interface for secure credential storage
type SecureStorage interface {
	Store(key, data string) error
	Retrieve(key string) (string, error)
	Delete(key string) error
}

// MacOSKeychain implements SecureStorage using macOS Keychain Services
type MacOSKeychain struct {
	serviceName string
}

// LinuxFileStorage implements SecureStorage using encrypted files
type LinuxFileStorage struct {
	serviceName string
}

// KeychainStorage provides cross-platform secure storage
type KeychainStorage struct {
	serviceName string
	impl        SecureStorage
}

// NewSecureStorage creates the appropriate secure storage implementation based on platform
func NewSecureStorage(serviceName string) SecureStorage {
	switch runtime.GOOS {
	case "darwin":
		return &MacOSKeychain{serviceName: serviceName}
	case "linux":
		return &LinuxFileStorage{serviceName: serviceName}
	default:
		return nil
	}
}

// NewKeychainStorage creates a new keychain storage instance
func NewKeychainStorage(serviceName string) *KeychainStorage {
	impl := NewSecureStorage(serviceName)
	return &KeychainStorage{
		serviceName: serviceName,
		impl:        impl,
	}
}

// Store saves data securely using the platform-specific implementation
func (k *KeychainStorage) Store(key, data string) error {
	if k.impl == nil {
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return k.impl.Store(key, data)
}

// Retrieve gets data securely using the platform-specific implementation
func (k *KeychainStorage) Retrieve(key string) (string, error) {
	if k.impl == nil {
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return k.impl.Retrieve(key)
}

// Delete removes data securely using the platform-specific implementation
func (k *KeychainStorage) Delete(key string) error {
	if k.impl == nil {
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return k.impl.Delete(key)
}

// MacOSKeychain implementation

// Store saves data in macOS Keychain
func (m *MacOSKeychain) Store(key, data string) error {
	cmd := exec.Command("security", "add-generic-password",
		"-U", // Update if exists
		"-s", m.serviceName,
		"-a", key,
		"-w", data)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to store in keychain: %w (output: %s)", err, string(output))
	}

	return nil
}

// Retrieve gets data from macOS Keychain
func (m *MacOSKeychain) Retrieve(key string) (string, error) {
	cmd := exec.Command("security", "find-generic-password",
		"-s", m.serviceName,
		"-a", key,
		"-w") // Return password only

	output, err := cmd.Output()
	if err != nil {
		if strings.Contains(err.Error(), "exit status 44") {
			return "", fmt.Errorf("key not found in keychain: %s", key)
		}
		return "", fmt.Errorf("failed to retrieve from keychain: %w", err)
	}

	data := strings.TrimSuffix(string(output), "\n")
	return data, nil
}

// Delete removes data from macOS Keychain
func (m *MacOSKeychain) Delete(key string) error {
	cmd := exec.Command("security", "delete-generic-password",
		"-s", m.serviceName,
		"-a", key)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(err.Error(), "exit status 44") {
			return nil
		}
		return fmt.Errorf("failed to delete from keychain: %w (output: %s)", err, string(output))
	}

	return nil
}

// LinuxFileStorage implementation

// Store saves data in encrypted file (Linux)
func (l *LinuxFileStorage) Store(key, data string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	credentialsDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(credentialsDir, 0o700); err != nil {
		return fmt.Errorf("failed to create credentials directory: %w", err)
	}

	filename := fmt.Sprintf(".%s_%s.json", l.serviceName, key)
	credentialsPath := filepath.Join(credentialsDir, filename)

	tempPath := credentialsPath + ".tmp"
	if err := os.WriteFile(tempPath, []byte(data), 0o600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	if err := os.Rename(tempPath, credentialsPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to replace credentials file: %w", err)
	}

	return nil
}

// Retrieve gets data from encrypted file (Linux)
func (l *LinuxFileStorage) Retrieve(key string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	filename := fmt.Sprintf(".%s_%s.json", l.serviceName, key)
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

// Delete removes data from encrypted file (Linux)
func (l *LinuxFileStorage) Delete(key string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	filename := fmt.Sprintf(".%s_%s.json", l.serviceName, key)
	credentialsPath := filepath.Join(home, ".claude", filename)

	err = os.Remove(credentialsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to delete credentials file: %w", err)
	}

	return nil
}
