// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/bvk/past/git"
	"github.com/bvk/past/gpg"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/xerrors"
)

var insertCmd = &cobra.Command{
	Use:   "insert [flags] <password-file>",
	Short: "Inserts a password to the in a new password-file.",
	RunE:  cmdInsert,
}

func init() {
	flags := insertCmd.Flags()
	flags.String("user", "", "Username to save along with the password.")
}

func cmdInsert(cmd *cobra.Command, args []string) (status error) {
	flags := cmd.Flags()
	if len(args) == 0 {
		return xerrors.Errorf("search string argument is required: %w", os.ErrInvalid)
	}
	if len(args) > 1 {
		return xerrors.Errorf("too many search string arguments: %w", os.ErrInvalid)
	}
	file := filepath.Join("./", args[0]+".gpg")
	dataDir, err := flags.GetString("data-dir")
	if err != nil {
		return xerrors.Errorf("could not get --data-dir value: %w", err)
	}
	if len(dataDir) == 0 {
		return xerrors.Errorf("data directory path be empty: %w", os.ErrInvalid)
	}
	user, err := flags.GetString("user")
	if err != nil {
		return xerrors.Errorf("could not get --user value: %w", err)
	}

	store, err := git.NewDir(dataDir)
	if err != nil {
		return xerrors.Errorf("could not create git directory instance: %w", err)
	}
	keyring, err := gpg.NewKeyring("")
	if err != nil {
		return xerrors.Errorf("could not create gpg key ring instance: %w", err)
	}

	keys, err := getFileKeys(store, keyring, file)
	if err != nil {
		return xerrors.Errorf("could not get keys for file %q: %w", file, err)
	}

	// Read the password without echo.
	passwd1, err := getPassword("Password:")
	if err != nil {
		return xerrors.Errorf("could not read password: %w", err)
	}
	passwd2, err := getPassword("Retype password:")
	if err != nil {
		return xerrors.Errorf("could not read retyped password: %w", err)
	}
	if passwd1 != passwd2 {
		return xerrors.Errorf("passwords do not match: %w", os.ErrInvalid)
	}

	lines := []string{passwd1}
	if len(user) > 0 {
		lines = append(lines, fmt.Sprintf("user: %s", user))
	}

	decrypted := []byte(strings.Join(lines, "\n") + "\n")
	encrypted, err := keyring.Encrypt(decrypted, keys)
	if err != nil {
		return xerrors.Errorf("could not encrypt new password: %w", err)
	}

	if err := store.AddFile(file, encrypted, os.FileMode(0644)); err != nil {
		return xerrors.Errorf("could not add file %q in git repo: %w", file, err)
	}
	defer func() {
		if status != nil {
			if err := store.Reset("HEAD"); err != nil {
				log.Panicf("could not undo adding file %q: %v", file, err)
				return
			}
		}
	}()

	msg := fmt.Sprintf("Added new password file %q", file)
	if err := store.Commit(msg); err != nil {
		return xerrors.Errorf("could not commit reinitialize change: %w", err)
	}
	return nil
}

func getPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", xerrors.Errorf("could not read password: %w", err)
	}
	return strings.TrimSpace(string(bytePassword)), nil
}
