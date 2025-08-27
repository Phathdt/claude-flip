package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

const version = "0.1.0"

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
	if alias != "" {
		fmt.Printf("Adding current account with alias: %s\n", alias)
	} else {
		fmt.Println("Adding current Claude Code account...")
	}
	fmt.Println("TODO: Implement add account functionality")
	return nil
}

func listAccounts(c *cli.Context) error {
	verbose := c.Bool("verbose")
	if verbose {
		fmt.Println("Listing accounts with detailed information...")
	} else {
		fmt.Println("Listing managed accounts...")
	}
	fmt.Println("TODO: Implement list accounts functionality")
	return nil
}

func switchAccount(c *cli.Context) error {
	target := c.Args().First()
	confirm := c.Bool("confirm")
	force := c.Bool("force")

	if target != "" {
		fmt.Printf("Switching to account: %s\n", target)
	} else {
		fmt.Println("Switching to next account in sequence...")
	}

	if confirm {
		fmt.Println("Confirmation enabled")
	}
	if force {
		fmt.Println("Force mode enabled (skipping safety checks)")
	}

	fmt.Println("TODO: Implement switch account functionality")
	return nil
}

func removeAccount(c *cli.Context) error {
	target := c.Args().First()
	if target == "" {
		return fmt.Errorf("account identifier required")
	}
	fmt.Printf("Removing account: %s\n", target)
	fmt.Println("TODO: Implement remove account functionality")
	return nil
}

func currentAccount(c *cli.Context) error {
	fmt.Println("Current active account:")
	fmt.Println("TODO: Implement current account functionality")
	return nil
}

func renameAccount(c *cli.Context) error {
	if c.Args().Len() < 2 {
		return fmt.Errorf("both account identifier and new alias required")
	}
	target := c.Args().Get(0)
	newAlias := c.Args().Get(1)
	fmt.Printf("Renaming account %s to alias: %s\n", target, newAlias)
	fmt.Println("TODO: Implement rename account functionality")
	return nil
}

func validateAccounts(c *cli.Context) error {
	fmt.Println("Validating all stored accounts...")
	fmt.Println("TODO: Implement validate accounts functionality")
	return nil
}
