// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bvk/past/git"
	"github.com/bvk/past/gpg"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

var showCmd = &cobra.Command{
	Use:   "show [flags] <password-file>",
	Short: "Decrypts a password-file and prints it's content.",
	RunE:  cmdShow,
}

func init() {
	flags := showCmd.Flags()
	flags.Uint32("line", 0, "When non-zero, only prints the data at given line.")
}

func cmdShow(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	if len(args) == 0 {
		return xerrors.Errorf("password file argument is required: %w", os.ErrInvalid)
	}
	if len(args) > 1 {
		return xerrors.Errorf("too many arguments: %w", os.ErrInvalid)
	}
	file := filepath.Join("./", args[0]+".gpg")
	dataDir, err := flags.GetString("data-dir")
	if err != nil {
		return xerrors.Errorf("could not get --data-dir value: %w", err)
	}
	if len(dataDir) == 0 {
		return xerrors.Errorf("data directory path be empty: %w", os.ErrInvalid)
	}
	line, err := flags.GetUint32("line")
	if err != nil {
		return xerrors.Errorf("could not get --line value: %w", err)
	}
	store, err := git.NewDir(dataDir)
	if err != nil {
		return xerrors.Errorf("could not create git directory instance: %w", err)
	}
	keyring, err := gpg.NewKeyring("")
	if err != nil {
		return xerrors.Errorf("could not create gpg key ring instance: %w", err)
	}
	data, err := store.ReadFile(file)
	if err != nil {
		return xerrors.Errorf("could not read file %q: %w", file, err)
	}
	decrypted, err := keyring.Decrypt(data)
	if err != nil {
		return xerrors.Errorf("could not decrypt file %q: %w", file, err)
	}
	if line == 0 {
		fmt.Printf("%s", decrypted)
		return nil
	}

	lines := bytes.Split(decrypted, []byte("\n"))
	if l := len(lines); line > uint32(l) {
		return xerrors.Errorf("file %q doesn't have line %d: %w", line, os.ErrInvalid)
	}
	fmt.Printf("%s\n", lines[int(line)-1])
	return nil
}
