// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Prints the names of all password-files.",
	RunE:  cmdList,
}

func cmdList(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	ps, err := newPasswordStore(flags)
	if err != nil {
		return xerrors.Errorf("could not create password store instance: %w", err)
	}

	files, err := ps.ListFiles()
	if err != nil {
		return xerrors.Errorf("could not list files in the git directory: %w", err)
	}
	for _, file := range files {
		fmt.Println(file)
	}
	return nil
}
