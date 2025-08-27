package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"claude-flip/internal/auth"
)

// Storage interface defines the contract for account storage
type Storage interface {
	SaveAccounts(accounts []*auth.Account) error
	LoadAccounts() ([]*auth.Account, error)
	DeleteAccount(accountID string) error
	Clear() error
}

// StorageManager handles platform-specific storage
type StorageManager struct {
	storage Storage
}

// NewStorageManager creates a new storage manager for the current platform
func NewStorageManager() (*StorageManager, error) {
	var storage Storage
	var err error

	switch runtime.GOOS {
	case "darwin":
		storage, err = NewKeychainStorage()
	case "linux":
		storage, err = NewFileStorage()
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	return &StorageManager{storage: storage}, nil
}

// SaveAccounts saves accounts to storage
func (sm *StorageManager) SaveAccounts(accounts []*auth.Account) error {
	return sm.storage.SaveAccounts(accounts)
}

// LoadAccounts loads accounts from storage
func (sm *StorageManager) LoadAccounts() ([]*auth.Account, error) {
	return sm.storage.LoadAccounts()
}

// DeleteAccount removes an account from storage
func (sm *StorageManager) DeleteAccount(accountID string) error {
	return sm.storage.DeleteAccount(accountID)
}

// Clear removes all accounts from storage
func (sm *StorageManager) Clear() error {
	return sm.storage.Clear()
}

// FileStorage implements Storage using encrypted files (Linux)
type FileStorage struct {
	dataDir string
}

// NewFileStorage creates a new file-based storage
func NewFileStorage() (*FileStorage, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	dataDir := filepath.Join(home, ".claude-flip")

	// Create directory with secure permissions
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &FileStorage{dataDir: dataDir}, nil
}

// SaveAccounts saves accounts to encrypted file
func (fs *FileStorage) SaveAccounts(accounts []*auth.Account) error {
	if accounts == nil {
		accounts = []*auth.Account{}
	}

	data, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal accounts: %w", err)
	}

	// For now, we'll use basic encryption (can be enhanced later)
	encryptedData, err := encrypt(data)
	if err != nil {
		return fmt.Errorf("failed to encrypt accounts data: %w", err)
	}

	accountsPath := filepath.Join(fs.dataDir, "accounts.enc")

	// Create backup if file exists
	if _, err := os.Stat(accountsPath); err == nil {
		backupPath := accountsPath + ".backup"
		if err := copyFile(accountsPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Write atomically using temporary file
	tempPath := accountsPath + ".tmp"
	if err := os.WriteFile(tempPath, encryptedData, 0o600); err != nil {
		return fmt.Errorf("failed to write accounts file: %w", err)
	}

	if err := os.Rename(tempPath, accountsPath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to replace accounts file: %w", err)
	}

	return nil
}

// LoadAccounts loads accounts from encrypted file
func (fs *FileStorage) LoadAccounts() ([]*auth.Account, error) {
	accountsPath := filepath.Join(fs.dataDir, "accounts.enc")

	if _, err := os.Stat(accountsPath); os.IsNotExist(err) {
		return []*auth.Account{}, nil // Return empty slice if file doesn't exist
	}

	encryptedData, err := os.ReadFile(accountsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read accounts file: %w", err)
	}

	data, err := decrypt(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt accounts data: %w", err)
	}

	var accounts []*auth.Account
	if err := json.Unmarshal(data, &accounts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal accounts: %w", err)
	}

	return accounts, nil
}

// DeleteAccount removes a specific account (implemented by reloading, filtering, and saving)
func (fs *FileStorage) DeleteAccount(accountID string) error {
	accounts, err := fs.LoadAccounts()
	if err != nil {
		return err
	}

	// Filter out the account to delete
	var filteredAccounts []*auth.Account
	for _, account := range accounts {
		if account.ID != accountID {
			filteredAccounts = append(filteredAccounts, account)
		}
	}

	return fs.SaveAccounts(filteredAccounts)
}

// Clear removes all accounts
func (fs *FileStorage) Clear() error {
	accountsPath := filepath.Join(fs.dataDir, "accounts.enc")
	if err := os.Remove(accountsPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove accounts file: %w", err)
	}
	return nil
}

// KeychainStorage implements Storage using macOS Keychain
type KeychainStorage struct {
	service string
}

// NewKeychainStorage creates a new keychain-based storage
func NewKeychainStorage() (*KeychainStorage, error) {
	return &KeychainStorage{
		service: "claude-flip",
	}, nil
}

// SaveAccounts saves accounts to keychain (placeholder - requires keychain implementation)
func (ks *KeychainStorage) SaveAccounts(accounts []*auth.Account) error {
	// For now, fall back to file storage on macOS
	// TODO: Implement proper keychain integration
	fileStorage, err := NewFileStorage()
	if err != nil {
		return err
	}
	return fileStorage.SaveAccounts(accounts)
}

// LoadAccounts loads accounts from keychain (placeholder)
func (ks *KeychainStorage) LoadAccounts() ([]*auth.Account, error) {
	// For now, fall back to file storage on macOS
	// TODO: Implement proper keychain integration
	fileStorage, err := NewFileStorage()
	if err != nil {
		return nil, err
	}
	return fileStorage.LoadAccounts()
}

// DeleteAccount removes an account from keychain (placeholder)
func (ks *KeychainStorage) DeleteAccount(accountID string) error {
	// For now, fall back to file storage on macOS
	// TODO: Implement proper keychain integration
	fileStorage, err := NewFileStorage()
	if err != nil {
		return err
	}
	return fileStorage.DeleteAccount(accountID)
}

// Clear removes all accounts from keychain (placeholder)
func (ks *KeychainStorage) Clear() error {
	// For now, fall back to file storage on macOS
	// TODO: Implement proper keychain integration
	fileStorage, err := NewFileStorage()
	if err != nil {
		return err
	}
	return fileStorage.Clear()
}

// copyFile creates a copy of a file
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}
