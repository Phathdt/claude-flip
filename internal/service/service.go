package service

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"claude-flip/internal/profile"
)

// Service provides the main business logic for Claude Flip
type Service struct {
	switcher *profile.Switcher
}

// NewService creates a new service instance
func NewService() (*Service, error) {
	switcher, err := profile.NewSwitcher()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize switcher: %w", err)
	}

	return &Service{
		switcher: switcher,
	}, nil
}

// ProfileInfo represents profile information for the CLI
type ProfileInfo struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Alias        string `json:"alias,omitempty"`
	AccountUuid  string `json:"account_uuid"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	LastActiveAt string `json:"last_active_at,omitempty"`
}

// AddCurrentAccount adds the current Claude Code account to managed profiles
func (s *Service) AddCurrentAccount(alias string) (*ProfileInfo, error) {
	// Generate profile name - use alias if provided, otherwise use email
	var profileName string
	if alias != "" {
		profileName = alias
	} else {
		// We'll let the switcher generate a name based on the email
		profileName = ""
	}

	// Save current account as profile
	profile, err := s.switcher.SaveCurrentAccount(profileName, alias)
	if err != nil {
		return nil, fmt.Errorf("failed to save current account: %w", err)
	}

	// Set this profile as the active one (since it's the current account)
	if err := s.switcher.SetActiveProfile(profile.Name); err != nil {
		return nil, fmt.Errorf("failed to set active profile: %w", err)
	}

	// Convert to ProfileInfo
	return s.profileToInfo(profile, true), nil
}

// ListAccounts returns all managed profiles
func (s *Service) ListProfiles() ([]*ProfileInfo, error) {
	profiles, err := s.switcher.ListProfiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	// Get active profile
	activeProfile, _ := s.switcher.GetCurrentActiveProfile()
	activeProfileName := ""
	if activeProfile != nil {
		activeProfileName = activeProfile.Name
	}

	// Convert to ProfileInfo slice
	var profileInfos []*ProfileInfo
	for _, profile := range profiles {
		isActive := profile.Name == activeProfileName
		profileInfos = append(profileInfos, s.profileToInfo(profile, isActive))
	}

	return profileInfos, nil
}

// GetCurrentAccount returns the currently active profile
func (s *Service) GetCurrentAccount() (*ProfileInfo, error) {
	profile, err := s.switcher.GetCurrentActiveProfile()
	if err != nil {
		return nil, fmt.Errorf("no active profile found: %w", err)
	}

	return s.profileToInfo(profile, true), nil
}

// SwitchToAccount switches to a specific profile
func (s *Service) SwitchToAccount(identifier string, force bool) error {
	if !force {
		if err := s.checkClaudeCodeNotRunning(); err != nil {
			return err
		}
	}

	// Switch to the target profile
	_, err := s.switcher.SwitchToAccount(identifier)
	if err != nil {
		return fmt.Errorf("failed to switch to profile: %w", err)
	}

	return nil
}

// RemoveAccount removes a profile from management
func (s *Service) RemoveAccount(identifier string) error {
	return s.switcher.DeleteProfile(identifier)
}

// RenameAccount changes the name/alias of a profile
func (s *Service) RenameAccount(identifier, newAlias string) error {
	return s.switcher.RenameProfile(identifier, "", newAlias)
}

// ValidateAccounts validates all stored profiles
func (s *Service) ValidateAccounts() map[string]error {
	profiles, err := s.switcher.ListProfiles()
	if err != nil {
		return map[string]error{
			"list_error": err,
		}
	}

	errors := make(map[string]error)
	for _, profile := range profiles {
		if err := s.switcher.ValidateProfile(profile.Name); err != nil {
			displayName := profile.Alias
			if displayName == "" {
				displayName = profile.Email
			}
			errors[displayName] = err
		}
	}

	return errors
}

// GetAccountByIdentifier gets a profile by identifier (for internal use)
func (s *Service) GetAccountByIdentifier(identifier string) (*ProfileInfo, error) {
	profiles, err := s.switcher.ListProfiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	// Get active profile for comparison
	activeProfile, _ := s.switcher.GetCurrentActiveProfile()
	activeProfileName := ""
	if activeProfile != nil {
		activeProfileName = activeProfile.Name
	}

	// Find matching profile
	for _, profile := range profiles {
		if profile.Name == identifier || profile.Email == identifier || profile.Alias == identifier {
			isActive := profile.Name == activeProfileName
			return s.profileToInfo(profile, isActive), nil
		}
	}

	return nil, fmt.Errorf("profile not found: %s", identifier)
}

// profileToInfo converts a profile.Profile to ProfileInfo
func (s *Service) profileToInfo(p *profile.Profile, isActive bool) *ProfileInfo {
	info := &ProfileInfo{
		Name:        p.Name,
		Email:       p.Email,
		Alias:       p.Alias,
		AccountUuid: p.AccountUuid,
		IsActive:    isActive,
		CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if !p.LastActiveAt.IsZero() {
		info.LastActiveAt = p.LastActiveAt.Format("2006-01-02 15:04:05")
	}

	return info
}

// checkClaudeCodeNotRunning checks if Claude Code is currently running
func (s *Service) checkClaudeCodeNotRunning() error {
	var processNames []string

	switch runtime.GOOS {
	case "darwin":
		processNames = []string{"Claude Code", "claude-code"}
	case "linux":
		processNames = []string{"claude-code"}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	for _, processName := range processNames {
		if isProcessRunning(processName) {
			return fmt.Errorf("Claude Code is currently running. Please close it before switching accounts")
		}
	}

	return nil
}

// isProcessRunning checks if a process is currently running
func isProcessRunning(processName string) bool {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pgrep", "-f", processName)
	case "linux":
		cmd = exec.Command("pgrep", "-f", processName)
	default:
		return false
	}

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) != ""
}
