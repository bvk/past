// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/bvk/past/msg"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

func main() {
	if err := doMain(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func doMain(ctx context.Context) error {
	flags := mainCmd.PersistentFlags()
	dataDir := filepath.Join(os.Getenv("HOME"), ".password-store")
	flags.String("data-dir", dataDir, "Data directory for the password store.")

	// If this program is invoked as chrome extension, dispatch to the native
	// messaging handler.
	if len(os.Args) == 2 {
		for _, extID := range ExtensionIDs {
			extensionArg := fmt.Sprintf("chrome-extension://%s", extID)
			if strings.HasPrefix(os.Args[1], extensionArg) {
				return runExtensionCmd(dataDir)
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

func runExtensionCmd(dataDir string) (status error) {
	// Redirect the logs to a file.
	file := filepath.Join(os.TempDir(), fmt.Sprintf("past-%s.log", os.Getenv("USER")))
	logfile, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(0600))
	if err != nil {
		return xerrors.Errorf("could not open log file %q: %w", file, err)
	}
	log.SetOutput(logfile)
	defer func() {
		if e := recover(); e != nil {
			log.Printf("%s: %s", e, debug.Stack())
		}
	}()

	// When invoked by chrome, PATH may not be the same, so fix it to a known
	// good defaults. For example, gpg command from GPGTools package in Mac OS X
	// is installed into /usr/local/bin directory, but it is not part of the PATH
	// when chrome extension invokes this program.
	homeBin := filepath.Join(os.Getenv("HOME"), "bin")
	os.Setenv("PATH", homeBin+":/usr/local/bin:/usr/bin:/bin")

	handler, err := msg.New(dataDir)
	if err != nil {
		return xerrors.Errorf("could not create native messaging handler: %w", err)
	}
	return handler.Serve(context.Background(), os.Stdin, os.Stdout)
}
