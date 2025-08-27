package auth

import (
	"fmt"
	"time"
)

// Account represents a Claude Code account
type Account struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name,omitempty"`
	Alias        string    `json:"alias,omitempty"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	UserID       string    `json:"user_id,omitempty"`
	ExpiresAt    int64     `json:"expires_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsActive     bool      `json:"is_active"`
}

// AccountManager manages multiple Claude Code accounts
type AccountManager struct {
	accounts map[string]*Account
}

// NewAccountManager creates a new account manager
func NewAccountManager() *AccountManager {
	return &AccountManager{
		accounts: make(map[string]*Account),
	}
}

// AddAccount adds a new account to the manager
func (am *AccountManager) AddAccount(account *Account) error {
	if account == nil {
		return fmt.Errorf("account cannot be nil")
	}

	if account.Email == "" {
		return fmt.Errorf("account email is required")
	}

	if account.AccessToken == "" {
		return fmt.Errorf("account access token is required")
	}

	// Generate ID if not provided
	if account.ID == "" {
		account.ID = generateAccountID(account.Email)
	}

	// Set timestamps
	now := time.Now()
	if account.CreatedAt.IsZero() {
		account.CreatedAt = now
	}
	account.UpdatedAt = now

	am.accounts[account.ID] = account
	return nil
}

// GetAccount retrieves an account by ID or email
func (am *AccountManager) GetAccount(identifier string) (*Account, error) {
	// Try by ID first
	if account, exists := am.accounts[identifier]; exists {
		return account, nil
	}

	// Try by email
	for _, account := range am.accounts {
		if account.Email == identifier {
			return account, nil
		}
	}

	return nil, fmt.Errorf("account not found: %s", identifier)
}

// GetAccountByAlias retrieves an account by alias
func (am *AccountManager) GetAccountByAlias(alias string) (*Account, error) {
	for _, account := range am.accounts {
		if account.Alias == alias {
			return account, nil
		}
	}
	return nil, fmt.Errorf("account with alias '%s' not found", alias)
}

// ListAccounts returns all accounts
func (am *AccountManager) ListAccounts() []*Account {
	accounts := make([]*Account, 0, len(am.accounts))
	for _, account := range am.accounts {
		accounts = append(accounts, account)
	}
	return accounts
}

// RemoveAccount removes an account by ID or email
func (am *AccountManager) RemoveAccount(identifier string) error {
	account, err := am.GetAccount(identifier)
	if err != nil {
		return err
	}

	delete(am.accounts, account.ID)
	return nil
}

// SetActiveAccount marks an account as active and others as inactive
func (am *AccountManager) SetActiveAccount(identifier string) error {
	activeAccount, err := am.GetAccount(identifier)
	if err != nil {
		return err
	}

	// Set all accounts to inactive
	for _, account := range am.accounts {
		account.IsActive = false
		account.UpdatedAt = time.Now()
	}

	// Set the specified account as active
	activeAccount.IsActive = true
	activeAccount.UpdatedAt = time.Now()

	return nil
}

// GetActiveAccount returns the currently active account
func (am *AccountManager) GetActiveAccount() (*Account, error) {
	for _, account := range am.accounts {
		if account.IsActive {
			return account, nil
		}
	}
	return nil, fmt.Errorf("no active account found")
}

// UpdateAccount updates an existing account
func (am *AccountManager) UpdateAccount(account *Account) error {
	if account == nil {
		return fmt.Errorf("account cannot be nil")
	}

	existing, err := am.GetAccount(account.ID)
	if err != nil {
		return fmt.Errorf("account not found for update: %w", err)
	}

	// Update fields
	existing.Email = account.Email
	existing.Name = account.Name
	existing.Alias = account.Alias
	existing.AccessToken = account.AccessToken
	existing.RefreshToken = account.RefreshToken
	existing.UserID = account.UserID
	existing.ExpiresAt = account.ExpiresAt
	existing.UpdatedAt = time.Now()

	return nil
}

// RenameAccount changes the alias of an account
func (am *AccountManager) RenameAccount(identifier, newAlias string) error {
	account, err := am.GetAccount(identifier)
	if err != nil {
		return err
	}

	// Check if alias is already in use
	if newAlias != "" {
		for _, acc := range am.accounts {
			if acc.ID != account.ID && acc.Alias == newAlias {
				return fmt.Errorf("alias '%s' is already in use", newAlias)
			}
		}
	}

	account.Alias = newAlias
	account.UpdatedAt = time.Now()

	return nil
}

// ValidateAccount checks if an account has valid credentials
func (am *AccountManager) ValidateAccount(identifier string) error {
	account, err := am.GetAccount(identifier)
	if err != nil {
		return err
	}

	if account.AccessToken == "" {
		return fmt.Errorf("account %s has no access token", account.Email)
	}

	// Check if token is expired
	if account.ExpiresAt > 0 && time.Now().Unix() > account.ExpiresAt {
		return fmt.Errorf("account %s token has expired", account.Email)
	}

	return nil
}

// GetNextAccount returns the next account in sequence for switching
func (am *AccountManager) GetNextAccount() (*Account, error) {
	accounts := am.ListAccounts()
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no accounts available")
	}

	if len(accounts) == 1 {
		return accounts[0], nil
	}

	// Find current active account
	currentIndex := -1
	for i, account := range accounts {
		if account.IsActive {
			currentIndex = i
			break
		}
	}

	// If no active account or it's the last one, return first account
	if currentIndex == -1 || currentIndex == len(accounts)-1 {
		return accounts[0], nil
	}

	// Return next account
	return accounts[currentIndex+1], nil
}

// AccountCount returns the number of managed accounts
func (am *AccountManager) AccountCount() int {
	return len(am.accounts)
}

// generateAccountID creates a unique ID for an account based on email
func generateAccountID(email string) string {
	return fmt.Sprintf("account_%s_%d", email, time.Now().Unix())
}

// GetDisplayName returns the display name for an account (alias or email)
func (a *Account) GetDisplayName() string {
	if a.Alias != "" {
		return a.Alias
	}
	return a.Email
}

// IsExpired checks if the account's token is expired
func (a *Account) IsExpired() bool {
	if a.ExpiresAt <= 0 {
		return false // No expiration set
	}
	return time.Now().Unix() > a.ExpiresAt
}
