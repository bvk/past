// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"os"

	"github.com/bvk/past"
	"github.com/bvk/past/git"
	"github.com/bvk/past/gpg"

	"github.com/spf13/pflag"
	"golang.org/x/xerrors"
)

func newPasswordStore(flags *pflag.FlagSet) (*past.PasswordStore, error) {
	dataDir, err := flags.GetString("data-dir")
	if err != nil {
		return nil, xerrors.Errorf("could not get --data-dir value: %w", err)
	}
	if len(dataDir) == 0 {
		return nil, xerrors.Errorf("data directory path be empty: %w", os.ErrInvalid)
	}
	repo, err := git.NewDir(dataDir)
	if err != nil {
		return nil, xerrors.Errorf("could not create git directory instance: %w", err)
	}
	keyring, err := gpg.NewKeyring("")
	if err != nil {
		return nil, xerrors.Errorf("could not create gpg key ring instance: %w", err)
	}
	return past.New(repo, keyring)
}
