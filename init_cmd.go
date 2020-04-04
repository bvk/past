// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/bvk/past/git"
	"github.com/bvk/past/gpg"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

var initCmd = &cobra.Command{
	Use:   "init [flags] <gpg-key>...",
	Short: "Creates or re-encrypts a password-store with GPG keys.",
	RunE:  cmdInit,
}

func init() {
	flags := initCmd.Flags()
	flags.Bool("skip-decrypt-failures", false, "When true, files that could not be decrypted will be skipped.")
}

func cmdInit(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	if len(args) == 0 {
		return xerrors.Errorf("at least one GPG id argument is required: %w", os.ErrInvalid)
	}
	dataDir, err := flags.GetString("data-dir")
	if err != nil {
		return xerrors.Errorf("could not get --data-dir value: %w", err)
	}

	create := false
	if _, err := os.Stat(dataDir); err != nil {
		create = true
	}

	if create {
		if err := doCreate(cmd, args); err != nil {
			return xerrors.Errorf("could not create password store: %w", err)
		}
	} else {
		if err := doReinit(cmd, args); err != nil {
			return xerrors.Errorf("could not re-initialize password store: %w", err)
		}
	}
	return nil
}

func doCreate(cmd *cobra.Command, args []string) (status error) {
	flags := cmd.Flags()
	dataDir, err := flags.GetString("data-dir")
	if err != nil {
		return xerrors.Errorf("could not get --data-dir value: %w", err)
	}
	keyring, err := gpg.NewKeyring("")
	if err != nil {
		return xerrors.Errorf("could not create gpg key ring instance: %w", err)
	}

	// Args should be encryptable key-ids or key-fingerprints or email ids
	// associated with encryptable key-ids that are not expired yet.
	now := time.Now()
	pks := keyring.PublicKeys()
	var keys []*gpg.PublicKey
	for _, arg := range args {
		var key *gpg.PublicKey
		for _, pk := range pks {
			if pk.CanEncrypt && (pk.ExpiresAt.IsZero() || pk.ExpiresAt.After(now)) {
				if strings.EqualFold(pk.KeyID, arg) || strings.EqualFold(pk.Fingerprint, arg) ||
					strings.EqualFold(pk.UserEmail, arg) {
					key = pk
					break
				}
			}
		}
		if key == nil {
			return xerrors.Errorf("could not find a valid gpg key for %q: %w", arg, os.ErrInvalid)
		}
		keys = append(keys, key)
	}

	initCmd := exec.Command("git", "init", dataDir)
	if err := initCmd.Run(); err != nil {
		return xerrors.Errorf("could not initialize git repository at %q: %w", dataDir, err)
	}
	defer func() {
		if status != nil {
			if err := os.RemoveAll(dataDir); err != nil {
				log.Panicf("could not undo creating directory %q: %v", dataDir, err)
				return
			}
		}
	}()

	store, err := git.NewDir(dataDir)
	if err != nil {
		return xerrors.Errorf("could not create git directory instance: %w", err)
	}

	fps := []string{}
	for _, key := range keys {
		fps = append(fps, key.Fingerprint)
	}
	data := []byte(strings.Join(fps, "\n") + "\n")
	if err := store.AddFile(".gpg-id", data, os.FileMode(0644)); err != nil {
		return xerrors.Errorf("could not create .gpg-id file with the key ids: %w", err)
	}
	msg := fmt.Sprintf("Created password store with keys %q", fps)
	if err := store.Commit(msg); err != nil {
		return xerrors.Errorf("could not perform initial commit: %w", err)
	}
	return nil
}

func doReinit(cmd *cobra.Command, args []string) (status error) {
	flags := cmd.Flags()
	dataDir, err := flags.GetString("data-dir")
	if err != nil {
		return xerrors.Errorf("could not get --data-dir value: %w", err)
	}
	skipDecryptFailures, err := flags.GetBool("skip-decrypt-failures")
	if err != nil {
		return xerrors.Errorf("could not get --skip-decrypt-failures value: %w", err)
	}

	store, err := git.NewDir(dataDir)
	if err != nil {
		return xerrors.Errorf("could not create git directory instance: %w", err)
	}
	keyring, err := gpg.NewKeyring("")
	if err != nil {
		return xerrors.Errorf("could not create gpg key ring instance: %w", err)
	}
	// Args should be encryptable key-ids or key-fingerprints or email ids
	// associated with encryptable key-ids that are not expired yet.
	now := time.Now()
	pks := keyring.PublicKeys()
	var keys []*gpg.PublicKey
	for _, arg := range args {
		var key *gpg.PublicKey
		for _, pk := range pks {
			if pk.CanEncrypt && (pk.ExpiresAt.IsZero() || pk.ExpiresAt.After(now)) {
				if strings.EqualFold(pk.KeyID, arg) || strings.EqualFold(pk.Fingerprint, arg) ||
					strings.EqualFold(pk.UserEmail, arg) {
					key = pk
					break
				}
			}
		}
		if key == nil {
			return xerrors.Errorf("could not find a valid gpg key for %q: %w", arg, os.ErrInvalid)
		}
		keys = append(keys, key)
	}

	gitFiles, err := store.ListFiles()
	if err != nil {
		return xerrors.Errorf("could not list files in the git directory: %w", err)
	}

	// Scan all git files to identify .gpg-id files for each directory.
	dirKeysMap := make(map[string][]string)
	for _, path := range gitFiles {
		base := filepath.Base(path)
		if base == ".gpg-id" {
			data, err := store.ReadFile(path)
			if err != nil {
				return xerrors.Errorf("could not read file %q: %w", path, err)
			}
			dirKeysMap[filepath.Dir(path)] = strings.Fields(string(data))
		}
	}
	if _, ok := dirKeysMap["."]; !ok {
		return xerrors.Errorf("could not find .gpg-id file in the git root directory: %w", os.ErrInvalid)
	}

	// Scan all git files to identify the encrypted .gpg files and their encryption key ids.
	fileKeyDirMap := make(map[string]string)
	for _, path := range gitFiles {
		if strings.HasSuffix(path, ".gpg") {
			keyDir := "."
			for d := filepath.Dir(path); d != "."; d = filepath.Dir(d) {
				if _, ok := dirKeysMap[d]; ok {
					keyDir = d
					break
				}
			}
			fileKeyDirMap[path] = keyDir
		}
	}

	// Decrypt and re-encrypt all files using the default keys.
	nupdated := 0
	defer func() {
		if nupdated > 0 && status != nil {
			if err := store.Reset("HEAD"); err != nil {
				log.Printf("could not reset repo changes on a failure (needs manual cleanup): %v", err)
			}
		}
	}()

	skipped := []string{}
	for path, dir := range fileKeyDirMap {
		if dir != "." {
			continue
		}
		data, err := store.ReadFile(path)
		if err != nil {
			return xerrors.Errorf("could not read file %q: %w", path, err)
		}
		decrypted, err := keyring.Decrypt(data)
		if err != nil {
			if !skipDecryptFailures {
				return xerrors.Errorf("could not decrypt file %q: %w", path, err)
			}
			skipped = append(skipped, path)
		}
		encrypted, err := keyring.Encrypt(decrypted, keys)
		if err != nil {
			return xerrors.Errorf("could not rencrypt %q with new keys: %w", path, err)
		}
		if err := store.UpdateFile(path, encrypted); err != nil {
			return xerrors.Errorf("could not update file %q in git repo: %w", path, err)
		}
		log.Printf("re-encrypted %q with new keys", path)
		nupdated++
	}

	if nupdated == 0 && len(fileKeyDirMap) > 0 {
		return xerrors.Errorf("could not decrypt any file: %w", err)
	}

	// Update the .gpg-id file in root dir with th key fingerprints.
	fps := []string{}
	for _, key := range keys {
		fps = append(fps, key.Fingerprint)
	}
	if err := store.UpdateFile(".gpg-id", []byte(strings.Join(fps, "\n")+"\n")); err != nil {
		return xerrors.Errorf("could not update .gpg-id file with new keys: %w", err)
	}

	msg := fmt.Sprintf("Reinitialized the repo with new keys %q", args)
	if err := store.Commit(msg); err != nil {
		return xerrors.Errorf("could not commit reinitialize change: %w", err)
	}

	if len(skipped) > 0 {
		log.Printf("warning: could not decrypt files %q, so they are skipped", skipped)
	}
	return nil
}
