package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/phathdt/claude-flip/internal/storage"
)

// ClaudeConfig represents the complete Claude Code configuration structure
// We store it as a raw JSON map to preserve all fields exactly as they are
type ClaudeConfig map[string]interface{}

// OAuthAccount contains OAuth authentication information from Claude Code
type OAuthAccount struct {
	AccountUuid      string `json:"accountUuid,omitempty"`
	EmailAddress     string `json:"emailAddress,omitempty"`
	OrganizationUuid string `json:"organizationUuid,omitempty"`
	OrganizationRole string `json:"organizationRole,omitempty"`
	WorkspaceRole    string `json:"workspaceRole,omitempty"`
	OrganizationName string `json:"organizationName,omitempty"`
}

// Credentials represents the structure of ~/.claude/.credentials.json
type Credentials struct {
	ClaudeAiOauth struct {
		AccessToken      string   `json:"accessToken"`
		RefreshToken     string   `json:"refreshToken"`
		ExpiresAt        int64    `json:"expiresAt"`
		Scopes           []string `json:"scopes"`
		SubscriptionType string   `json:"subscriptionType"`
	} `json:"claudeAiOauth"`
}

// AuthConfig contains authentication information
type AuthConfig struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	Email        string `json:"email,omitempty"`
	ExpiresAt    int64  `json:"expires_at,omitempty"`
}

// UserConfig contains user preferences and settings
type UserConfig struct {
	Email    string                 `json:"email,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// FindClaudeConfigDir locates the Claude Code configuration directory
func FindClaudeConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Claude Code can store config in multiple locations, return home directory
	// as the LoadClaudeConfig function will handle finding the actual file
	return home, nil
}

// LoadClaudeConfig reads and parses the Claude Code configuration
func LoadClaudeConfig() (*ClaudeConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Try different possible locations and file names for Claude Code config
	configPaths := []string{
		filepath.Join(home, ".claude.json"),
		filepath.Join(home, ".claude", ".claude.json"),
		filepath.Join(home, ".claude", "claude.json"),
		filepath.Join(home, ".claude", "config.json"),
	}

	var config ClaudeConfig
	var lastErr error

	// Load main config file - now we load the COMPLETE config as a map
	for _, configPath := range configPaths {
		data, err := os.ReadFile(configPath)
		if err != nil {
			lastErr = err
			continue
		}

		config = make(ClaudeConfig)
		if err := json.Unmarshal(data, &config); err != nil {
			lastErr = fmt.Errorf("failed to parse config file %s: %w", configPath, err)
			continue
		}
		break
	}

	if config == nil {
		return nil, fmt.Errorf("no valid Claude Code config file found: %w", lastErr)
	}

	// Load credentials using platform-specific method
	if credentials, err := loadCredentialsForConfig(); err == nil {
		// Store credentials in a special field for our use
		config["_cflip_credentials"] = *credentials
	}

	return &config, nil
}

// SaveClaudeConfig writes the configuration back to disk
func SaveClaudeConfig(config *ClaudeConfig) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Claude config is stored at ~/.claude.json
	configPath := filepath.Join(home, ".claude.json")

	// Create backup before modifying
	if _, err := os.Stat(configPath); err == nil {
		backupPath := configPath + ".backup"
		if err := copyFile(configPath, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Create a clean copy without our internal fields
	cleanConfig := make(ClaudeConfig)
	for key, value := range *config {
		if !strings.HasPrefix(key, "_cflip_") {
			cleanConfig[key] = value
		}
	}

	data, err := json.MarshalIndent(cleanConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write atomically using temporary file
	tempPath := configPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write temporary config file: %w", err)
	}

	if err := os.Rename(tempPath, configPath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to replace config file: %w", err)
	}

	return nil
}

// ValidateConfig checks if the configuration contains required fields
func ValidateConfig(config ClaudeConfig) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	oauthAccount, ok := config["oauthAccount"].(map[string]interface{})
	if !ok || oauthAccount == nil {
		return fmt.Errorf("no OAuth account information found")
	}

	email, ok := oauthAccount["emailAddress"].(string)
	if !ok || email == "" {
		return fmt.Errorf("no email found in configuration")
	}

	accountUuid, ok := oauthAccount["accountUuid"].(string)
	if !ok || accountUuid == "" {
		return fmt.Errorf("no account UUID found in configuration")
	}

	return nil
}

// GetUserEmail extracts the user email from config
func (c ClaudeConfig) GetUserEmail() string {
	if oauthAccount, ok := c["oauthAccount"].(map[string]interface{}); ok {
		if email, ok := oauthAccount["emailAddress"].(string); ok {
			return email
		}
	}
	return ""
}

// GetAccountUuid extracts the account UUID from config
func (c ClaudeConfig) GetAccountUuid() string {
	if oauthAccount, ok := c["oauthAccount"].(map[string]interface{}); ok {
		if uuid, ok := oauthAccount["accountUuid"].(string); ok {
			return uuid
		}
	}
	return ""
}

// GetOrganizationName extracts the organization name from config
func (c ClaudeConfig) GetOrganizationName() string {
	if oauthAccount, ok := c["oauthAccount"].(map[string]interface{}); ok {
		if name, ok := oauthAccount["organizationName"].(string); ok {
			return name
		}
	}
	return ""
}

// GetCredentials extracts stored credentials from config
func (c ClaudeConfig) GetCredentials() (*Credentials, bool) {
	if credsData, ok := c["_cflip_credentials"]; ok {
		// Handle both direct Credentials struct and map[string]interface{} cases
		switch v := credsData.(type) {
		case Credentials:
			return &v, true
		case map[string]interface{}:
			// Convert map to Credentials struct
			data, err := json.Marshal(v)
			if err != nil {
				return nil, false
			}
			var creds Credentials
			if err := json.Unmarshal(data, &creds); err != nil {
				return nil, false
			}
			return &creds, true
		}
	}
	return nil, false
}

// SetOAuthAccount updates the oauthAccount section in the config
func (c ClaudeConfig) SetOAuthAccount(oauthData map[string]interface{}) {
	c["oauthAccount"] = oauthData
}

// loadCredentialsForConfig loads credentials using platform-specific method
func loadCredentialsForConfig() (*Credentials, error) {
	// Use the SecureStorage Capture method to read from Claude Code's native storage
	storage := storage.NewSecureStorage()
	credentialsJSON, err := storage.Capture()
	if err != nil {
		return nil, fmt.Errorf("failed to capture credentials: %w", err)
	}

	var credentials Credentials
	if err := json.Unmarshal([]byte(credentialsJSON), &credentials); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return &credentials, nil
}

// copyFile creates a copy of a file
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o600)
}
