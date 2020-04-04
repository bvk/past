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

	dirPubKeysMap map[string][]*gpg.PublicKey
}

func NewPasswordStore(store *git.Dir, keyring *gpg.Keyring) (*PasswordStore, error) {
	gitFiles, err := store.ListFiles()
	if err != nil {
		return nil, xerrors.Errorf("could not list files in the git directory: %w", err)
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
		dirPubKeysMap: dirPubKeysMap,
	}
	return ps, nil
}

func (ps *PasswordStore) AddPasswordFile(name, password string, rest [][2]string) (status error) {
	file := filepath.Join("./", name+".gpg")
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
	return nil
}

func (ps *PasswordStore) UpdatePasswordFile(name, password string, rest [][2]string) (status error) {
	file := filepath.Join("./", name+".gpg")
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
