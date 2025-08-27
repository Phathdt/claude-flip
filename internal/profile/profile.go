package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/phathdt/claude-flip/internal/config"
)

// Profile represents a saved Claude Code account configuration
type Profile struct {
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Alias        string    `json:"alias,omitempty"`
	AccountUuid  string    `json:"account_uuid"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastActiveAt time.Time `json:"last_active_at,omitempty"`

	// Claude Code configuration data
	ClaudeConfig *config.ClaudeConfig `json:"claude_config"`
	Credentials  *config.Credentials  `json:"credentials"`
}

// ProfileManager manages Claude Code account profiles
type ProfileManager struct {
	profilesDir string
	configPath  string
}

// Config represents the cflip configuration
type Config struct {
	ActiveProfile string            `json:"active_profile,omitempty"`
	Profiles      map[string]string `json:"profiles"` // profile_name -> email mapping
	LastUpdated   time.Time         `json:"last_updated"`
}

// NewProfileManager creates a new profile manager
func NewProfileManager() (*ProfileManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	profilesDir := filepath.Join(home, ".cflip")
	configPath := filepath.Join(profilesDir, "config.json")

	// Create the profiles directory if it doesn't exist
	if err := os.MkdirAll(profilesDir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create profiles directory: %w", err)
	}

	return &ProfileManager{
		profilesDir: profilesDir,
		configPath:  configPath,
	}, nil
}

// SaveProfile saves a profile to disk
func (pm *ProfileManager) SaveProfile(profile *Profile) error {
	if profile.Name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Generate filename based on email (sanitized)
	filename := sanitizeFilename(profile.Email) + ".profile"
	profilePath := filepath.Join(pm.profilesDir, filename)

	profile.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	// Write atomically using temporary file
	tempPath := profilePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write profile file: %w", err)
	}

	if err := os.Rename(tempPath, profilePath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to replace profile file: %w", err)
	}

	// Update the main config
	return pm.updateConfig(profile.Name, profile.Email)
}

// LoadProfile loads a profile from disk
func (pm *ProfileManager) LoadProfile(identifier string) (*Profile, error) {
	profilePath, err := pm.findProfilePath(identifier)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile file: %w", err)
	}

	var profile Profile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	return &profile, nil
}

// ListProfiles returns all available profiles
func (pm *ProfileManager) ListProfiles() ([]*Profile, error) {
	entries, err := os.ReadDir(pm.profilesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	var profiles []*Profile
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".profile" {
			profilePath := filepath.Join(pm.profilesDir, entry.Name())

			data, err := os.ReadFile(profilePath)
			if err != nil {
				continue // Skip invalid files
			}

			var profile Profile
			if err := json.Unmarshal(data, &profile); err != nil {
				continue // Skip invalid files
			}

			profiles = append(profiles, &profile)
		}
	}

	return profiles, nil
}

// DeleteProfile removes a profile from disk
func (pm *ProfileManager) DeleteProfile(identifier string) error {
	profilePath, err := pm.findProfilePath(identifier)
	if err != nil {
		return err
	}

	// Load profile to get name for config cleanup
	profile, err := pm.LoadProfile(identifier)
	if err != nil {
		return fmt.Errorf("failed to load profile for deletion: %w", err)
	}

	// Remove the profile file
	if err := os.Remove(profilePath); err != nil {
		return fmt.Errorf("failed to remove profile file: %w", err)
	}

	// Update config to remove profile reference
	config, err := pm.LoadConfig()
	if err != nil {
		return err
	}

	delete(config.Profiles, profile.Name)
	if config.ActiveProfile == profile.Name {
		config.ActiveProfile = ""
	}

	return pm.SaveConfig(config)
}

// GetActiveProfile returns the currently active profile
func (pm *ProfileManager) GetActiveProfile() (*Profile, error) {
	config, err := pm.LoadConfig()
	if err != nil {
		return nil, err
	}

	if config.ActiveProfile == "" {
		return nil, fmt.Errorf("no active profile set")
	}

	return pm.LoadProfile(config.ActiveProfile)
}

// SetActiveProfile marks a profile as active
func (pm *ProfileManager) SetActiveProfile(identifier string) error {
	// Verify the profile exists
	profile, err := pm.LoadProfile(identifier)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	config, err := pm.LoadConfig()
	if err != nil {
		return err
	}

	config.ActiveProfile = profile.Name
	config.LastUpdated = time.Now()

	return pm.SaveConfig(config)
}

// LoadConfig loads the main cflip configuration
func (pm *ProfileManager) LoadConfig() (*Config, error) {
	if _, err := os.Stat(pm.configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return &Config{
			Profiles:    make(map[string]string),
			LastUpdated: time.Now(),
		}, nil
	}

	data, err := os.ReadFile(pm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if config.Profiles == nil {
		config.Profiles = make(map[string]string)
	}

	return &config, nil
}

// SaveConfig saves the main cflip configuration
func (pm *ProfileManager) SaveConfig(config *Config) error {
	config.LastUpdated = time.Now()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write atomically using temporary file
	tempPath := pm.configPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if err := os.Rename(tempPath, pm.configPath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to replace config file: %w", err)
	}

	return nil
}

// findProfilePath finds the profile file path by name or email
func (pm *ProfileManager) findProfilePath(identifier string) (string, error) {
	// First try by sanitized email filename
	filename := sanitizeFilename(identifier) + ".profile"
	profilePath := filepath.Join(pm.profilesDir, filename)

	if _, err := os.Stat(profilePath); err == nil {
		return profilePath, nil
	}

	// Search all profiles for matching name or email
	profiles, err := pm.ListProfiles()
	if err != nil {
		return "", err
	}

	for _, profile := range profiles {
		if profile.Name == identifier || profile.Email == identifier {
			filename := sanitizeFilename(profile.Email) + ".profile"
			return filepath.Join(pm.profilesDir, filename), nil
		}
	}

	return "", fmt.Errorf("profile not found: %s", identifier)
}

// updateConfig updates the main config with profile information
func (pm *ProfileManager) updateConfig(name, email string) error {
	config, err := pm.LoadConfig()
	if err != nil {
		return err
	}

	config.Profiles[name] = email
	return pm.SaveConfig(config)
}

// sanitizeFilename sanitizes a string to be safe for use as a filename
func sanitizeFilename(s string) string {
	// Replace unsafe characters with underscores
	result := ""
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' {
			result += string(r)
		} else {
			result += "_"
		}
	}
	return result
}
