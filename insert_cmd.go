// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/bvk/past/store"

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
	ps, err := newPasswordStore(flags)
	if err != nil {
		return xerrors.Errorf("could not create password-store instance: %w", err)
	}

	if len(args) == 0 {
		return xerrors.Errorf("search string argument is required: %w", os.ErrInvalid)
	}
	if len(args) > 1 {
		return xerrors.Errorf("too many search string arguments: %w", os.ErrInvalid)
	}
	file := filepath.Join("./", args[0])

	user, err := flags.GetString("user")
	if err != nil {
		return xerrors.Errorf("could not get --user value: %w", err)
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

	vs := store.NewValues(nil)
	if len(user) > 0 {
		vs.Set("username", user)
	}

	data := store.Format(passwd1, vs.Bytes())
	if err := ps.CreateFile(file, data, os.FileMode(0644)); err != nil {
		return xerrors.Errorf("could not insert new file %q: %w", file, err)
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
