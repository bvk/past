// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

var editCmd = &cobra.Command{
	Use:   "edit [flags] <password-file>",
	Short: "Updates an existing password-file with external editor.",
	RunE:  cmdEdit,
}

func init() {
	flags := editCmd.Flags()
	flags.String("editor", os.Getenv("EDITOR"), "Name of the editor.")
}

func cmdEdit(cmd *cobra.Command, args []string) (status error) {
	flags := cmd.Flags()
	ps, err := newPasswordStore(flags)
	if err != nil {
		return xerrors.Errorf("could not create password store instance: %w", err)
	}

	if len(args) == 0 {
		return xerrors.Errorf("password file argument is required: %w", os.ErrInvalid)
	}
	if len(args) > 1 {
		return xerrors.Errorf("too many arguments: %w", os.ErrInvalid)
	}
	file := filepath.Join("./", args[0])

	editor, err := flags.GetString("editor")
	if err != nil {
		return xerrors.Errorf("could not get --editor value: %w", err)
	}
	if len(editor) == 0 {
		return xerrors.Errorf("editor name cannot be empty: %w", os.ErrInvalid)
	}

	oldDecrypted, err := ps.ReadFile(file)
	if err != nil {
		return xerrors.Errorf("could not read password file %q: %w", file, err)
	}

	temp, err := ioutil.TempFile("", "file")
	if err != nil {
		return xerrors.Errorf("could not create temporary file: %w", err)
	}
	defer func() {
		temp.Close()
		if err := os.Remove(temp.Name()); err != nil {
			log.Panicf("could not remove temporary file %q with potentially decrypted data: %v", temp.Name(), err)
		}
	}()

	if _, err := temp.Write(oldDecrypted); err != nil {
		return xerrors.Errorf("could not write to temporary file: %w", err)
	}

	editorCmd := exec.Command(editor, temp.Name())
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	if err := editorCmd.Run(); err != nil {
		return xerrors.Errorf("editor returned with non-zero exit status: %w", err)
	}

	newDecrypted, err := ioutil.ReadFile(temp.Name())
	if err != nil {
		return xerrors.Errorf("could not read temporary file: %w", err)
	}
	if v := bytes.Compare(oldDecrypted, newDecrypted); v == 0 {
		log.Printf("no changes detected, so file is not updated")
		return nil
	}

	if err := ps.UpdateFile(file, newDecrypted); err != nil {
		return xerrors.Errorf("could not update file %q: %w", file, err)
	}
	return nil
}
