// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	if err := doMain(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func doMain(ctx context.Context) error {
	flags := mainCmd.PersistentFlags()
	flags.String("data-dir", filepath.Join(os.Getenv("HOME"), ".password-store"),
		"Data directory for the password store.")

	// If this program is invoked by chrome extension, just execute the chrome handler.
	if len(os.Args) == 2 {
		for _, extID := range ExtensionIDs {
			extensionArg := fmt.Sprintf("chrome-extension://%s", extID)
			if strings.HasPrefix(os.Args[1], extensionArg) {
				return cmdChrome(flags, os.Args[1:])
			}
		}
	}

	mainCmd.AddCommand(initCmd)
	mainCmd.AddCommand(listCmd)
	mainCmd.AddCommand(editCmd)
	mainCmd.AddCommand(scanCmd)
	mainCmd.AddCommand(insertCmd)
	mainCmd.AddCommand(generateCmd)
	mainCmd.AddCommand(gitCmd)
	mainCmd.AddCommand(keysCmd)
	mainCmd.AddCommand(showCmd)
	mainCmd.AddCommand(installCmd)
	mainCmd.AddCommand(importCmd)

	mainCmd.SilenceUsage = true
	mainCmd.SilenceErrors = true
	if _, err := mainCmd.ExecuteC(); err != nil {
		return err
	}
	return nil
}

var mainCmd = &cobra.Command{
	Use:   "past subcmd [flags]",
	Short: "Manages GPG encrypted password-files in a Git repository.",
}
