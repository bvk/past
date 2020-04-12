// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bvk/past/git"
	"github.com/bvk/past/gpg"
	"golang.org/x/xerrors"
)

type PasswordStore struct {
	store   *git.Dir
	keyring *gpg.Keyring

	gitFiles []string

	dirPubKeysMap map[string][]*gpg.PublicKey
}

func Create(store *git.Dir, keyring *gpg.Keyring, fingerprints []string) (_ *PasswordStore, status error) {
	if store == nil {
		return nil, xerrors.Errorf("git repository cannot be nil: %w", os.ErrInvalid)
	}
	if keyring == nil {
		return nil, xerrors.Errorf("keyring cannot be nil: %w", os.ErrInvalid)
	}

	file := "./.gpg-id"
	content := strings.Join(fingerprints, "\n")
	if err := store.AddFile(file, []byte(content), os.FileMode(0644)); err != nil {
		return nil, xerrors.Errorf("could not add file %q in git repo: %w", file, err)
	}
	defer func() {
		if status != nil {
			if err := store.Reset("HEAD"); err != nil {
				log.Panicf("could not undo adding file %q: %v", file, err)
				return
			}
		}
	}()

	msg := fmt.Sprintf("Initialized password store with keys %q", fingerprints)
	if err := store.Commit(msg); err != nil {
		return nil, xerrors.Errorf("could not commit gpg keys file: %w", err)
	}

	return NewPasswordStore(store, keyring)
}

func NewPasswordStore(store *git.Dir, keyring *gpg.Keyring) (*PasswordStore, error) {
	if store == nil {
		return nil, xerrors.Errorf("git repository cannot be nil: %w", os.ErrInvalid)
	}
	if keyring == nil {
		return nil, xerrors.Errorf("keyring cannot be nil: %w", os.ErrInvalid)
	}

	gitFiles, err := store.ListFiles()
	if err != nil {
		return nil, xerrors.Errorf("could not list files in the git directory: %w", err)
	}
	for i, file := range gitFiles {
		gitFiles[i] = filepath.Clean(filepath.Join("./", file))
	}

	// Scan all git files to identify .gpg-id files for each directory.
	dirKeysMap := make(map[string][]string)
	for _, path := range gitFiles {
		base := filepath.Base(path)
		if base == ".gpg-id" {
			data, err := store.ReadFile(path)
			if err != nil {
				return nil, xerrors.Errorf("could not read file %q: %w", path, err)
			}
			dirKeysMap[filepath.Dir(path)] = strings.Fields(string(data))
		}
	}
	if _, ok := dirKeysMap["."]; !ok {
		return nil, xerrors.Errorf("could not find .gpg-id file in the git root directory: %w", os.ErrInvalid)
	}

	dirPubKeysMap := make(map[string][]*gpg.PublicKey)
	for dir, keys := range dirKeysMap {
		var pkeys []*gpg.PublicKey
		for _, pk := range keyring.PublicKeys() {
			for _, key := range keys {
				if pk.Fingerprint == key {
					pkeys = append(pkeys, pk)
				}
			}
		}
		if len(pkeys) != len(keys) {
			log.Printf("warning: gpg keyring doesn't have all required keys for directory %q: %w", dir, os.ErrInvalid)
			continue
		}
		dirPubKeysMap[dir] = pkeys
	}

	if _, ok := dirPubKeysMap["."]; !ok {
		return nil, xerrors.Errorf("could not find default GPG keys for the repository: %w", os.ErrInvalid)
	}

	ps := &PasswordStore{
		store:         store,
		keyring:       keyring,
		gitFiles:      gitFiles,
		dirPubKeysMap: dirPubKeysMap,
	}
	return ps, nil
}

func (ps *PasswordStore) DefaultKeys() []*gpg.PublicKey {
	return ps.dirPubKeysMap["."]
}

func (ps *PasswordStore) AddPasswordFile(name, password string, rest [][2]string) (status error) {
	file := filepath.Clean(filepath.Join("./", name+".gpg"))
	pkeys, err := ps.FileKeys(file)
	if err != nil {
		return xerrors.Errorf("could not find appropriate keys for file %q: %w", file, err)
	}

	lines := []string{password}
	for _, kv := range rest {
		lines = append(lines, fmt.Sprintf("%s: %s", kv[0], kv[1]))
	}
	decrypted := []byte(strings.Join(lines, "\n") + "\n")

	encrypted, err := ps.keyring.Encrypt(decrypted, pkeys)
	if err != nil {
		return xerrors.Errorf("could not encrypt new password: %w", err)
	}

	if err := ps.store.AddFile(file, encrypted, os.FileMode(0644)); err != nil {
		return xerrors.Errorf("could not add file %q in git repo: %w", file, err)
	}
	defer func() {
		if status != nil {
			if err := ps.store.Reset("HEAD"); err != nil {
				log.Panicf("could not undo adding file %q: %v", file, err)
				return
			}
		}
	}()

	msg := fmt.Sprintf("Added new password file %q", file)
	if err := ps.store.Commit(msg); err != nil {
		return xerrors.Errorf("could not commit add file change: %w", err)
	}
	ps.gitFiles = append(ps.gitFiles, file)
	return nil
}

func (ps *PasswordStore) UpdatePasswordFile(name, password string, rest [][2]string) (status error) {
	file := filepath.Clean(filepath.Join("./", name+".gpg"))
	pkeys, err := ps.FileKeys(file)
	if err != nil {
		return xerrors.Errorf("could not find appropriate keys for file %q: %w", file, err)
	}

	lines := []string{password}
	for _, kv := range rest {
		lines = append(lines, fmt.Sprintf("%s: %s", kv[0], kv[1]))
	}
	decrypted := []byte(strings.Join(lines, "\n") + "\n")

	encrypted, err := ps.keyring.Encrypt(decrypted, pkeys)
	if err != nil {
		return xerrors.Errorf("could not encrypt new password: %w", err)
	}

	if err := ps.store.UpdateFile(file, encrypted); err != nil {
		return xerrors.Errorf("could not update file %q in git repo: %w", file, err)
	}
	defer func() {
		if status != nil {
			if err := ps.store.Reset("HEAD"); err != nil {
				log.Panicf("could not undo updating file %q: %v", file, err)
				return
			}
		}
	}()

	msg := fmt.Sprintf("Updated password file %q", file)
	if err := ps.store.Commit(msg); err != nil {
		return xerrors.Errorf("could not commit update file change: %w", err)
	}
	return nil
}

func (ps *PasswordStore) ReplacePasswordFile(oldName, newName, password string, rest [][2]string) (status error) {
	oldFile := filepath.Clean(filepath.Join("./", oldName+".gpg"))
	if exists, err := ps.FileExists(oldFile); err != nil {
		return xerrors.Errorf("could not determine if old file %q exists: %w", oldName, err)
	} else if !exists {
		return xerrors.Errorf("old file %q doesn't exist: %w", oldName, os.ErrNotExist)
	}

	newFile := filepath.Clean(filepath.Join("./", newName+".gpg"))
	pkeys, err := ps.FileKeys(newFile)
	if err != nil {
		return xerrors.Errorf("could not find appropriate keys for file %q: %w", newName, err)
	}

	lines := []string{password}
	for _, kv := range rest {
		lines = append(lines, fmt.Sprintf("%s: %s", kv[0], kv[1]))
	}
	decrypted := []byte(strings.Join(lines, "\n") + "\n")

	encrypted, err := ps.keyring.Encrypt(decrypted, pkeys)
	if err != nil {
		return xerrors.Errorf("could not encrypt new password: %w", err)
	}

	if err := ps.store.AddFile(newFile, encrypted, os.FileMode(0644)); err != nil {
		return xerrors.Errorf("could not add file %q in git repo: %w", newFile, err)
	}
	defer func() {
		if status != nil {
			if err := ps.store.Reset("HEAD"); err != nil {
				log.Panicf("could not undo adding file %q: %v", newFile, err)
				return
			}
		}
	}()

	if err := ps.store.RemoveFile(oldFile); err != nil {
		return xerrors.Errorf("could not remove file %q: %w", oldFile, err)
	}

	msg := fmt.Sprintf("Replaced %q with %q.", oldFile, newFile)
	if err := ps.store.Commit(msg); err != nil {
		return xerrors.Errorf("could not commit add file change: %w", err)
	}

	var gitFiles []string
	for _, file := range ps.gitFiles {
		if file == oldFile {
			gitFiles = append(gitFiles, newFile)
		} else {
			gitFiles = append(gitFiles, file)
		}
	}
	ps.gitFiles = gitFiles
	return nil
}

func (ps *PasswordStore) FileExists(path string) (bool, error) {
	for _, file := range ps.gitFiles {
		if file == path {
			return true, nil
		}
	}
	return false, nil
}

func (ps *PasswordStore) FileKeys(path string) ([]*gpg.PublicKey, error) {
	keys := ps.dirPubKeysMap["."]
	for d := filepath.Dir(path); d != "."; d = filepath.Dir(d) {
		if ks, ok := ps.dirPubKeysMap[d]; ok {
			keys = ks
			break
		}
	}

	// Make a copy.
	var clone []*gpg.PublicKey
	for _, key := range keys {
		newKey := new(gpg.PublicKey)
		*newKey = *key
		clone = append(clone, newKey)
	}
	return clone, nil
}
