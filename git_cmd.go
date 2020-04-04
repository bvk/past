// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

var gitCmd = &cobra.Command{
	Use:   "git [flags] cmds [git-flags]",
	Short: "Runs git(1) command on the password-store repository.",
	RunE:  cmdGit,
}

func cmdGit(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	dataDir, err := flags.GetString("data-dir")
	if err != nil {
		return xerrors.Errorf("could not get --data-dir value: %w", err)
	}
	if len(dataDir) == 0 {
		return xerrors.Errorf("data directory path be empty: %w", os.ErrInvalid)
	}
	gitCmd := exec.Command("git", "-C", dataDir)
	gitCmd.Args = append(gitCmd.Args, args...)
	gitCmd.Stdin, gitCmd.Stdout, gitCmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return gitCmd.Run()
}
