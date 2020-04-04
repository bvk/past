// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bvk/past/git"
	"github.com/bvk/past/gpg"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

var keysCmd = &cobra.Command{
	Use:   "keys [flags]",
	Short: "Prints GPG public keys information.",
	RunE:  cmdKeys,
}

func init() {
	flags := keysCmd.Flags()
	flags.Bool("trusted", false, "When true, prints only trusted keys.")
	flags.Bool("unexpired", false, "When true, does not print expired keys.")
}

func cmdKeys(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	trusted, err := flags.GetBool("trusted")
	if err != nil {
		return xerrors.Errorf("could not get --trusted value: %w", err)
	}
	unexpired, err := flags.GetBool("unexpired")
	if err != nil {
		return xerrors.Errorf("could not get --unexpired value: %w", err)
	}

	keyring, err := gpg.NewKeyring("")
	if err != nil {
		return xerrors.Errorf("could not create keys key ring instance: %w", err)
	}

	now := time.Now()
	pks := keyring.PublicKeys()
	for _, pk := range pks {
		if trusted && !pk.Trusted {
			continue
		}
		if unexpired && (!pk.ExpiresAt.IsZero() && now.After(pk.ExpiresAt)) {
			continue
		}
		data, _ := json.MarshalIndent(pk, "", "  ")
		fmt.Printf("%s\n", data)
	}
	return nil
}

func getFileKeys(store *git.Dir, keyring *gpg.Keyring, path string) ([]*gpg.PublicKey, error) {
	files, err := store.ListFiles()
	if err != nil {
		return nil, xerrors.Errorf("could not list files in the git directory: %w", err)
	}
	// Scan all git files to identify .gpg-id files for each directory.
	keyDirs := make(map[string]struct{})
	for _, path := range files {
		base := filepath.Base(path)
		if base == ".gpg-id" {
			keyDirs[filepath.Dir(path)] = struct{}{}
		}
	}
	if _, ok := keyDirs["."]; !ok {
		return nil, xerrors.Errorf("could not find the default gpg keys: %w", os.ErrInvalid)
	}

	keys := []string{}
	for d := filepath.Dir(path); d != "."; d = filepath.Dir(d) {
		if _, ok := keyDirs[d]; ok {
			path := filepath.Join(d, ".gpg-id")
			data, err := store.ReadFile(path)
			if err != nil {
				return nil, xerrors.Errorf("could not read file %q: %w", path, err)
			}
			keys = strings.Fields(string(data))
			break
		}
	}
	if len(keys) == 0 {
		file := ".gpg-id"
		data, err := store.ReadFile(file)
		if err != nil {
			return nil, xerrors.Errorf("could not read file %q: %w", file, err)
		}
		keys = strings.Fields(string(data))
	}
	var pkeys []*gpg.PublicKey
	for _, pk := range keyring.PublicKeys() {
		for _, key := range keys {
			if pk.Fingerprint == key {
				pkeys = append(pkeys, pk)
			}
		}
	}
	if len(pkeys) == 0 {
		return nil, xerrors.Errorf("gpg keyring doesn't have the required keys %q: %w", keys, os.ErrInvalid)
	}
	return pkeys, nil
}
