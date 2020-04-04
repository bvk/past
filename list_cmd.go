// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bvk/past/git"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Prints the names of all password files.",
	RunE:  cmdList,
}

func cmdList(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	dataDir, err := flags.GetString("data-dir")
	if err != nil {
		return xerrors.Errorf("could not get --data-dir value: %w", err)
	}
	if len(dataDir) == 0 {
		return xerrors.Errorf("data directory path be empty: %w", os.ErrInvalid)
	}

	store, err := git.NewDir(dataDir)
	if err != nil {
		return xerrors.Errorf("could not create git directory instance: %w", err)
	}

	files, err := store.ListFiles()
	if err != nil {
		return xerrors.Errorf("could not list files in the git directory: %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file, ".gpg") {
			fmt.Println(strings.TrimSuffix(file, ".gpg"))
		}
	}
	return nil
}
