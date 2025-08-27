package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"claude-flip/internal/config"
	"claude-flip/internal/storage"
)

// Switcher handles switching between Claude Code accounts
type Switcher struct {
	profileManager *ProfileManager
}

// NewSwitcher creates a new account switcher
func NewSwitcher() (*Switcher, error) {
	pm, err := NewProfileManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create profile manager: %w", err)
	}

	return &Switcher{
		profileManager: pm,
	}, nil
}

// SaveCurrentAccount saves the current Claude Code account as a profile
// SaveCurrentAccount saves the current Claude Code account as a profile
func (s *Switcher) SaveCurrentAccount(name, alias string) (*Profile, error) {
	// Load current Claude Code configuration
	claudeConfig, err := config.LoadClaudeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load Claude Code configuration: %w", err)
	}

	// Validate the configuration
	if err := config.ValidateConfig(*claudeConfig); err != nil {
		return nil, fmt.Errorf("invalid Claude Code configuration: %w", err)
	}

	// Get credentials from the config (they're already loaded)
	credentials, ok := claudeConfig.GetCredentials()
	if !ok {
		return nil, fmt.Errorf("failed to get credentials from config")
	}

	// Use email as profile name if no name provided
	profileName := name
	if profileName == "" {
		profileName = claudeConfig.GetUserEmail()
	}

	// Create profile
	now := time.Now()
	profile := &Profile{
		Name:         profileName,
		Email:        claudeConfig.GetUserEmail(),
		Alias:        alias,
		AccountUuid:  claudeConfig.GetAccountUuid(),
		CreatedAt:    now,
		UpdatedAt:    now,
		LastActiveAt: now, // Since this is the current account, set as last active
		ClaudeConfig: claudeConfig,
		Credentials:  credentials,
	}

	// Save profile
	if err := s.profileManager.SaveProfile(profile); err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}

	return profile, nil
}

// SwitchToAccount switches to a specific account profile
// SwitchToAccount switches to a specific account profile
func (s *Switcher) SwitchToAccount(identifier string) (*Profile, error) {
	var targetProfile *Profile
	var err error

	if identifier == "" {
		// Switch to next profile in sequence
		targetProfile, err = s.GetNextProfile()
		if err != nil {
			return nil, fmt.Errorf("failed to get next profile: %w", err)
		}
	} else {
		// Load specific target profile
		targetProfile, err = s.profileManager.LoadProfile(identifier)
		if err != nil {
			return nil, fmt.Errorf("failed to load target profile: %w", err)
		}
	}

	// Before switching, save current account if it's not already saved
	currentEmail := ""
	if currentConfig, err := config.LoadClaudeConfig(); err == nil {
		currentEmail = currentConfig.GetUserEmail()
	}

	// Check if current account is already saved
	shouldSaveCurrentAccount := true
	if currentEmail != "" {
		if currentProfile, err := s.profileManager.LoadProfile(currentEmail); err == nil {
			// Update the existing profile with current state
			currentClaudeConfig, err := config.LoadClaudeConfig()
			if err != nil {
				return nil, fmt.Errorf("failed to load current Claude config for backup: %w", err)
			}

			currentCredentials, err := s.loadCredentials()
			if err != nil {
				return nil, fmt.Errorf("failed to load current credentials for backup: %w", err)
			}

			currentProfile.ClaudeConfig = currentClaudeConfig
			currentProfile.Credentials = currentCredentials

			if err := s.profileManager.SaveProfile(currentProfile); err != nil {
				return nil, fmt.Errorf("failed to update current profile: %w", err)
			}

			shouldSaveCurrentAccount = false
		}
	}

	if shouldSaveCurrentAccount && currentEmail != "" {
		// Auto-save current account with email as name
		if _, err := s.SaveCurrentAccount(currentEmail, ""); err != nil {
			// Log warning but don't fail the switch
			fmt.Printf("Warning: failed to backup current account: %v\n", err)
		}
	}

	// Apply target profile configuration
	if err := s.applyProfile(targetProfile); err != nil {
		return nil, fmt.Errorf("failed to apply target profile: %w", err)
	}

	// Mark as active
	if err := s.profileManager.SetActiveProfile(targetProfile.Name); err != nil {
		return nil, fmt.Errorf("failed to set active profile: %w", err)
	}

	return targetProfile, nil
}

// GetCurrentActiveProfile returns the currently active profile
func (s *Switcher) GetCurrentActiveProfile() (*Profile, error) {
	return s.profileManager.GetActiveProfile()
}

// ListProfiles returns all available profiles
func (s *Switcher) ListProfiles() ([]*Profile, error) {
	return s.profileManager.ListProfiles()
}

// DeleteProfile removes a profile
func (s *Switcher) DeleteProfile(identifier string) error {
	return s.profileManager.DeleteProfile(identifier)
}

// RenameProfile changes a profile's name/alias
func (s *Switcher) RenameProfile(identifier, newName, newAlias string) error {
	profile, err := s.profileManager.LoadProfile(identifier)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	if newName != "" {
		profile.Name = newName
	}
	profile.Alias = newAlias

	return s.profileManager.SaveProfile(profile)
}

// ValidateProfile checks if a profile has valid credentials
func (s *Switcher) ValidateProfile(identifier string) error {
	profile, err := s.profileManager.LoadProfile(identifier)
	if err != nil {
		return err
	}

	if profile.ClaudeConfig == nil {
		return fmt.Errorf("profile %s has no Claude configuration", profile.Name)
	}

	// Check for oauthAccount in the map
	oauthAccount, ok := (*profile.ClaudeConfig)["oauthAccount"].(map[string]interface{})
	if !ok || oauthAccount == nil {
		return fmt.Errorf("profile %s has no OAuth account information", profile.Name)
	}

	if profile.Credentials == nil || profile.Credentials.ClaudeAiOauth.AccessToken == "" {
		return fmt.Errorf("profile %s has no access token", profile.Name)
	}

	// TODO: Could add token expiration check here
	return nil
}

// SetActiveProfile marks a profile as active without switching Claude config
func (s *Switcher) SetActiveProfile(identifier string) error {
	return s.profileManager.SetActiveProfile(identifier)
}

// GetNextProfile returns the next profile in sequence for switching
func (s *Switcher) GetNextProfile() (*Profile, error) {
	profiles, err := s.profileManager.ListProfiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no profiles available")
	}

	if len(profiles) == 1 {
		return profiles[0], nil
	}

	// Get current active profile
	activeProfile, err := s.profileManager.GetActiveProfile()
	if err != nil {
		// No active profile, return first one
		return profiles[0], nil
	}

	// Find current active profile in the list
	currentIndex := -1
	for i, profile := range profiles {
		if profile.Name == activeProfile.Name {
			currentIndex = i
			break
		}
	}

	// If current profile not found or it's the last one, return first profile
	if currentIndex == -1 || currentIndex == len(profiles)-1 {
		return profiles[0], nil
	}

	// Return next profile
	return profiles[currentIndex+1], nil
}

// applyProfile applies a profile's configuration to Claude Code
func (s *Switcher) applyProfile(profile *Profile) error {
	if profile.ClaudeConfig == nil {
		return fmt.Errorf("profile has no Claude configuration")
	}

	if profile.Credentials == nil {
		return fmt.Errorf("profile has no credentials")
	}

	// Update the oauthAccount section with fresh credentials before saving
	if oauthAccount, ok := (*profile.ClaudeConfig)["oauthAccount"].(map[string]interface{}); ok {
		// We don't store credentials in the oauthAccount section, they go in a separate file
		// Just ensure the account info is preserved as-is
		(*profile.ClaudeConfig)["oauthAccount"] = oauthAccount
	}

	// Save the main Claude config
	if err := config.SaveClaudeConfig(profile.ClaudeConfig); err != nil {
		return fmt.Errorf("failed to save Claude config: %w", err)
	}

	// Save the credentials file
	if err := s.saveCredentials(profile.Credentials); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	return nil
}

// loadCredentials loads the Claude Code credentials
// LoadCredentials loads Claude Code credentials using platform-specific storage
func LoadCredentials() (*config.Credentials, error) {
	switch runtime.GOOS {
	case "darwin":
		return loadCredentialsMacOS()
	case "linux":
		return loadCredentialsLinux()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// SaveCredentials saves Claude Code credentials using platform-specific storage
func SaveCredentials(credentials *config.Credentials) error {
	switch runtime.GOOS {
	case "darwin":
		return saveCredentialsMacOS(credentials)
	case "linux":
		return saveCredentialsLinux(credentials)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// loadCredentialsMacOS loads credentials from macOS Keychain
func loadCredentialsMacOS() (*config.Credentials, error) {
	keychain := storage.NewKeychainStorage("Claude Code-credentials")
	
	// Try to get current user for account key
	user := os.Getenv("USER")
	if user == "" {
		user = "default"
	}
	
	data, err := keychain.Retrieve(user)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials from keychain: %w", err)
	}
	
	var credentials config.Credentials
	if err := json.Unmarshal([]byte(data), &credentials); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}
	
	return &credentials, nil
}

// saveCredentialsMacOS saves credentials to macOS Keychain
func saveCredentialsMacOS(credentials *config.Credentials) error {
	data, err := json.Marshal(credentials)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}
	
	keychain := storage.NewKeychainStorage("Claude Code-credentials")
	
	// Try to get current user for account key
	user := os.Getenv("USER")
	if user == "" {
		user = "default"
	}
	
	if err := keychain.Store(user, string(data)); err != nil {
		return fmt.Errorf("failed to store credentials in keychain: %w", err)
	}
	
	return nil
}

// loadCredentialsLinux loads credentials from file system
func loadCredentialsLinux() (*config.Credentials, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	credentialsPath := filepath.Join(home, ".claude", ".credentials.json")
	data, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	var credentials config.Credentials
	if err := json.Unmarshal(data, &credentials); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return &credentials, nil
}

// saveCredentialsLinux saves credentials to file system
func saveCredentialsLinux(credentials *config.Credentials) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	credentialsPath := filepath.Join(home, ".claude", ".credentials.json")

	data, err := json.MarshalIndent(credentials, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Write atomically using temporary file
	tempPath := credentialsPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	if err := os.Rename(tempPath, credentialsPath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to replace credentials file: %w", err)
	}

	return nil
}

func (s *Switcher) loadCredentials() (*config.Credentials, error) {
	return LoadCredentials()
}

// saveCredentials saves the Claude Code credentials
func (s *Switcher) saveCredentials(credentials *config.Credentials) error {
	return SaveCredentials(credentials)
}
