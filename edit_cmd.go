// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bvk/past/git"
	"github.com/bvk/past/gpg"

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
	editor, err := flags.GetString("editor")
	if err != nil {
		return xerrors.Errorf("could not get --editor value: %w", err)
	}
	if len(editor) == 0 {
		return xerrors.Errorf("editor name cannot be empty: %w", os.ErrInvalid)
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

	oldEncrypted, err := store.ReadFile(file)
	if err != nil {
		return xerrors.Errorf("could not read file %q: %w", file, err)
	}
	oldDecrypted, err := keyring.Decrypt(oldEncrypted)
	if err != nil {
		return xerrors.Errorf("could not decrypt file %q: %w", file, err)
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
		return nil
	}

	newEncrypted, err := keyring.Encrypt(newDecrypted, keys)
	if err != nil {
		return xerrors.Errorf("could not encrypt new password: %w", err)
	}

	if err := store.UpdateFile(file, newEncrypted); err != nil {
		return xerrors.Errorf("could not update file %q in git repo: %w", file, err)
	}
	defer func() {
		if status != nil {
			if err := store.Reset("HEAD"); err != nil {
				log.Panicf("could not undo adding file %q: %v", file, err)
				return
			}
		}
	}()

	msg := fmt.Sprintf("Edited password file %q with %q", file, editor)
	if err := store.Commit(msg); err != nil {
		return xerrors.Errorf("could not commit reinitialize change: %w", err)
	}
	return nil
}
