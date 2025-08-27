package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/phathdt/claude-flip/internal/logger"
	"github.com/phathdt/claude-flip/internal/service"

	"github.com/urfave/cli/v2"
)

const version = "0.1.0"

// setupLogging configures the logger based on CLI flags
func setupLogging(c *cli.Context) error {
	logLevelStr := c.String("log-level")
	logFormat := c.String("log-format")

	var logLevel logger.LogLevel
	switch strings.ToLower(logLevelStr) {
	case "debug":
		logLevel = logger.LevelDebug
	case "info":
		logLevel = logger.LevelInfo
	case "warn", "warning":
		logLevel = logger.LevelWarn
	case "error":
		logLevel = logger.LevelError
	default:
		logLevel = logger.LevelInfo
	}

	config := &logger.LogConfig{
		Level:      logLevel,
		Format:     logFormat,
		Output:     "stderr",
		AddSource:  logLevel == logger.LevelDebug,
		Structured: false,
	}

	log, err := logger.New(config)
	if err != nil {
		return fmt.Errorf("failed to setup logging: %w", err)
	}

	logger.SetDefault(log)
	return nil
}

func main() {
	app := &cli.App{
		Name:    "cflip",
		Usage:   "A fast CLI tool to manage and switch between multiple Claude Code accounts",
		Version: version,
		Authors: []*cli.Author{
			{
				Name:  "phathdt",
				Email: "phathdt379@gmail.com",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"l"},
				Usage:   "Set logging level (debug, info, warn, error)",
				Value:   "info",
				EnvVars: []string{"CFLIP_LOG_LEVEL"},
			},
			&cli.StringFlag{
				Name:    "log-format",
				Usage:   "Set logging format (text, json)",
				Value:   "text",
				EnvVars: []string{"CFLIP_LOG_FORMAT"},
			},
		},
		Before: func(c *cli.Context) error {
			return setupLogging(c)
		},
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "Add current Claude Code account to managed accounts",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "alias",
						Aliases: []string{"n"},
						Usage:   "Custom alias for the account",
					},
				},
				Action: addAccount,
			},
			{
				Name:    "list",
				Aliases: []string{"ls", "l"},
				Usage:   "List all managed accounts (shows which one is active)",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Show detailed account information",
					},
				},
				Action: listAccounts,
			},
			{
				Name:      "switch",
				Aliases:   []string{"sw", "s"},
				Usage:     "Switch to account (next in sequence if no argument provided)",
				ArgsUsage: "[account_number|email]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "confirm",
						Aliases: []string{"c"},
						Usage:   "Show confirmation prompt before switching",
					},
					&cli.BoolFlag{
						Name:  "force",
						Usage: "Force switch (skip safety checks)",
					},
				},
				Action: switchAccount,
			},
			{
				Name:      "remove",
				Aliases:   []string{"rm", "r"},
				Usage:     "Remove an account from management",
				ArgsUsage: "<account_number|email>",
				Action:    removeAccount,
			},
			{
				Name:    "current",
				Aliases: []string{"cur"},
				Usage:   "Show current active account",
				Action:  currentAccount,
			},
			{
				Name:      "rename",
				Usage:     "Rename account alias",
				ArgsUsage: "<account_number|email> <new_alias>",
				Action:    renameAccount,
			},
			{
				Name:   "validate",
				Usage:  "Validate all stored accounts",
				Action: validateAccounts,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func addAccount(c *cli.Context) error {
	alias := c.String("alias")

	svc, err := service.NewService()
	if err != nil {
		return fmt.Errorf("failed to initialize service: %w", err)
	}

	if alias != "" {
		logger.Progress("Adding current account with alias: %s", alias)
	} else {
		logger.Progress("Adding current Claude Code account...")
	}

	profile, err := svc.AddCurrentAccount(alias)
	if err != nil {
		return fmt.Errorf("failed to add account: %w", err)
	}

	displayName := profile.Alias
	if displayName == "" {
		displayName = profile.Email
	}

	logger.Success("Account added successfully: %s", displayName)
	if profile.Email != displayName {
		logger.Plain("   Email: %s", profile.Email)
	}

	// Log audit event
	log := logger.NewDefault()
	log.AccountAdded(profile.Email, profile.Alias)

	return nil
}

func listAccounts(c *cli.Context) error {
	verbose := c.Bool("verbose")

	svc, err := service.NewService()
	if err != nil {
		return fmt.Errorf("failed to initialize service: %w", err)
	}

	profiles, err := svc.ListProfiles()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if len(profiles) == 0 {
		logger.InfoMsg("No accounts found. Use 'cflip add' to add your first account.")
		return nil
	}

	logger.InfoMsg("📋 Managed accounts (%d):", len(profiles))
	logger.Plain("")

	for i, profile := range profiles {
		statusIcon := "○"
		if profile.IsActive {
			statusIcon = "●"
		}

		displayName := profile.Alias
		if displayName == "" {
			displayName = profile.Email
		}

		accountInfo := fmt.Sprintf("%s %d. %s", statusIcon, i+1, displayName)
		if profile.Email != displayName {
			accountInfo += fmt.Sprintf(" (%s)", profile.Email)
		}

		if profile.IsActive {
			accountInfo += " [ACTIVE]"
		}

		// Note: We don't have expiration check in ProfileInfo, could add if needed

		logger.Plain("%s", accountInfo)

		if verbose {
			logger.Plain("   Created: %s", profile.CreatedAt)
			logger.Plain("   Updated: %s", profile.UpdatedAt)
			if profile.LastActiveAt != "" {
				logger.Plain("   Last Active: %s", profile.LastActiveAt)
			}
			logger.Plain("")
		}
	}

	return nil
}

func switchAccount(c *cli.Context) error {
	target := c.Args().First()
	confirm := c.Bool("confirm")
	force := c.Bool("force")

	svc, err := service.NewService()
	if err != nil {
		return fmt.Errorf("failed to initialize service: %w", err)
	}

	// Get current account for audit logging
	var fromEmail string
	if currentAcc, err := svc.GetCurrentAccount(); err == nil {
		fromEmail = currentAcc.Email
	}

	// If target is numeric, convert to account by index
	if target != "" {
		if index, err := strconv.Atoi(target); err == nil && index > 0 {
			accounts, _ := svc.ListProfiles()
			if index <= len(accounts) {
				target = accounts[index-1].Email
			} else {
				return fmt.Errorf("invalid account number: %d (only %d accounts available)", index, len(accounts))
			}
		}
	}

	if target != "" {
		logger.Progress("Switching to account: %s", target)
	} else {
		logger.Progress("Switching to next account in sequence...")
	}

	if confirm && !force {
		logger.Question("Are you sure you want to switch accounts? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			logger.ErrorMsg("Switch cancelled")
			return nil
		}
	}

	err = svc.SwitchToAccount(target, force)
	if err != nil {
		return fmt.Errorf("failed to switch account: %w", err)
	}

	// Get the account we switched to
	currentAccount, err := svc.GetCurrentAccount()
	if err != nil {
		return fmt.Errorf("failed to get current account: %w", err)
	}

	displayName := currentAccount.Alias
	if displayName == "" {
		displayName = currentAccount.Email
	}
	logger.Success("Successfully switched to: %s", displayName)
	logger.InfoMsg("💡 Please restart Claude Code to use the new account")

	// Log audit event
	log := logger.NewDefault()
	log.AccountSwitched(fromEmail, currentAccount.Email)

	return nil
}

func removeAccount(c *cli.Context) error {
	target := c.Args().First()
	if target == "" {
		return fmt.Errorf("account identifier required")
	}

	svc, err := service.NewService()
	if err != nil {
		return fmt.Errorf("failed to initialize service: %w", err)
	}

	// If target is numeric, convert to account by index
	if index, err := strconv.Atoi(target); err == nil && index > 0 {
		accounts, _ := svc.ListProfiles()
		if index <= len(accounts) {
			target = accounts[index-1].Email
		} else {
			return fmt.Errorf("invalid account number: %d (only %d accounts available)", index, len(accounts))
		}
	}

	logger.Warning("🗑️  Removing account: %s", target)

	// Confirmation prompt
	logger.Question("Are you sure you want to remove this account? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		logger.ErrorMsg("Removal cancelled")
		return nil
	}

	err = svc.RemoveAccount(target)
	if err != nil {
		return fmt.Errorf("failed to remove account: %w", err)
	}

	logger.Success("Account removed successfully: %s", target)

	// Log audit event
	log := logger.NewDefault()
	log.AccountRemoved(target)

	return nil
}

func currentAccount(c *cli.Context) error {
	svc, err := service.NewService()
	if err != nil {
		return fmt.Errorf("failed to initialize service: %w", err)
	}

	profile, err := svc.GetCurrentAccount()
	if err != nil {
		return fmt.Errorf("no active account found: %w", err)
	}

	displayName := profile.Alias
	if displayName == "" {
		displayName = profile.Email
	}

	logger.InfoMsg("📍 Current active account:")
	logger.Plain("   Name: %s", displayName)
	logger.Plain("   Email: %s", profile.Email)
	if profile.AccountUuid != "" {
		logger.Plain("   User ID: %s", profile.AccountUuid)
	}
	logger.Plain("   Last Updated: %s", profile.UpdatedAt)

	// For now, we'll always show as ACTIVE since it's the current profile
	logger.Success("   Status: ACTIVE")

	return nil
}

func renameAccount(c *cli.Context) error {
	if c.Args().Len() < 2 {
		return fmt.Errorf("both account identifier and new alias required")
	}
	target := c.Args().Get(0)
	newAlias := c.Args().Get(1)

	svc, err := service.NewService()
	if err != nil {
		return fmt.Errorf("failed to initialize service: %w", err)
	}

	// Get old alias for audit logging
	var oldAlias string
	if acc, err := svc.GetAccountByIdentifier(target); err == nil {
		oldAlias = acc.Alias
	}

	// If target is numeric, convert to account by index
	if index, err := strconv.Atoi(target); err == nil && index > 0 {
		accounts, _ := svc.ListProfiles()
		if index <= len(accounts) {
			target = accounts[index-1].Email
		} else {
			return fmt.Errorf("invalid account number: %d (only %d accounts available)", index, len(accounts))
		}
	}

	logger.Progress("🏷️  Renaming account %s to alias: %s", target, newAlias)

	err = svc.RenameAccount(target, newAlias)
	if err != nil {
		return fmt.Errorf("failed to rename account: %w", err)
	}

	logger.Success("Account renamed successfully: %s", newAlias)

	// Log audit event
	log := logger.NewDefault()
	log.AccountRenamed(target, oldAlias, newAlias)

	return nil
}

func validateAccounts(c *cli.Context) error {
	svc, err := service.NewService()
	if err != nil {
		return fmt.Errorf("failed to initialize service: %w", err)
	}

	logger.Progress("🔍 Validating all stored accounts...")

	errors := svc.ValidateAccounts()
	if len(errors) == 0 {
		logger.Success("All accounts are valid")
		return nil
	}

	logger.ErrorMsg("Found %d invalid accounts:", len(errors))
	logger.Plain("")
	for accountName, err := range errors {
		logger.Plain("  • %s: %s", accountName, err.Error())
	}

	return fmt.Errorf("%d accounts failed validation", len(errors))
}
