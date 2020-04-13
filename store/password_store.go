// Copyright (c) 2020 BVK Chaitanya

package store

import (
	"bytes"
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
	if err := store.CreateFile(file, []byte(content), os.FileMode(0644)); err != nil {
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

	return New(store, keyring)
}

func New(store *git.Dir, keyring *gpg.Keyring) (*PasswordStore, error) {
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

func (ps *PasswordStore) ListFiles() ([]string, error) {
	var files []string
	for _, file := range ps.gitFiles {
		if strings.HasSuffix(file, ".gpg") {
			files = append(files, strings.TrimSuffix(file, ".gpg"))
		}
	}
	return files, nil
}

// ReadFile returns a password file's content in unencrypted form.
func (ps *PasswordStore) ReadFile(path string) ([]byte, error) {
	file := filepath.Clean(filepath.Join("./", path+".gpg"))
	encrypted, err := ps.store.ReadFile(file)
	if err != nil {
		return nil, xerrors.Errorf("could not read file %q: %w", file, err)
	}
	decrypted, err := ps.keyring.Decrypt(encrypted)
	if err != nil {
		return nil, xerrors.Errorf("could not decrypt file %q: %w", file, err)
	}
	return decrypted, nil
}

// WriteFile creates or overwrites a password file with the input data. Input
// data is assumed to be in unencrypted form with a password and any optional
// user data prepared using the Format function. New password file will be
// created with the input mode if target file doesn't exist.
func (ps *PasswordStore) WriteFile(path string, data []byte, mode os.FileMode) error {
	file := filepath.Clean(filepath.Join("./", path+".gpg"))
	pkeys, err := ps.FileKeys(file)
	if err != nil {
		return xerrors.Errorf("could not find appropriate keys for file %q: %w", file, err)
	}
	encrypted, err := ps.keyring.Encrypt(data, pkeys)
	if err != nil {
		return xerrors.Errorf("could not encrypt password file data: %w", err)
	}

	msg := fmt.Sprintf("Created new password file %q.", file)
	if _, err := ps.store.Stat(file); err == nil {
		msg = fmt.Sprintf("Updated password file %q.", file)
	}
	cb := func() error {
		return ps.store.WriteFile(file, encrypted, mode)
	}
	if err := ps.store.Apply(msg, cb); err != nil {
		return xerrors.Errorf("could not write to password file %q: %w", file, err)
	}
	return nil
}

// UpdateFile is similar to WriteFile, but fails if target file doesn't exist.
func (ps *PasswordStore) UpdateFile(path string, data []byte) error {
	file := filepath.Clean(filepath.Join("./", path+".gpg"))
	pkeys, err := ps.FileKeys(file)
	if err != nil {
		return xerrors.Errorf("could not find appropriate keys for file %q: %w", file, err)
	}
	encrypted, err := ps.keyring.Encrypt(data, pkeys)
	if err != nil {
		return xerrors.Errorf("could not encrypt password file data: %w", err)
	}

	msg := fmt.Sprintf("Updated password file %q.", file)
	cb := func() error {
		return ps.store.UpdateFile(file, encrypted)
	}
	if err := ps.store.Apply(msg, cb); err != nil {
		return xerrors.Errorf("could not update password file %q: %w", file, err)
	}
	return nil
}

// CreateFile is similar to WriteFile, but fails if target file already exists.
func (ps *PasswordStore) CreateFile(path string, data []byte, mode os.FileMode) error {
	file := filepath.Clean(filepath.Join("./", path+".gpg"))
	pkeys, err := ps.FileKeys(file)
	if err != nil {
		return xerrors.Errorf("could not find appropriate keys for file %q: %w", file, err)
	}
	encrypted, err := ps.keyring.Encrypt(data, pkeys)
	if err != nil {
		return xerrors.Errorf("could not encrypt password file data: %w", err)
	}

	msg := fmt.Sprintf("Created new password file %q.", file)
	cb := func() error {
		return ps.store.CreateFile(file, encrypted, os.FileMode(0644))
	}
	if err := ps.store.Apply(msg, cb); err != nil {
		return xerrors.Errorf("could not create new password file %q: %w", file, err)
	}
	return nil
}

func (ps *PasswordStore) ReplaceFile(oldpath, newpath string, data []byte) error {
	oldfile := filepath.Clean(filepath.Join("./", oldpath+".gpg"))
	newfile := filepath.Clean(filepath.Join("./", newpath+".gpg"))
	pkeys, err := ps.FileKeys(newfile)
	if err != nil {
		return xerrors.Errorf("could not find appropriate keys for file %q: %w", newfile, err)
	}
	encrypted, err := ps.keyring.Encrypt(data, pkeys)
	if err != nil {
		return xerrors.Errorf("could not encrypt password file data: %w", err)
	}

	msg := fmt.Sprintf("Replaced file %q with %q.", oldfile, newfile)
	cb := func() error {
		if err := ps.store.Rename(oldfile, newfile); err != nil {
			return err
		}
		return ps.store.UpdateFile(newfile, encrypted)
	}
	if err := ps.store.Apply(msg, cb); err != nil {
		return xerrors.Errorf("could not create replace password file %q to %q: %w", oldfile, newfile, err)
	}
	return nil
}

func (ps *PasswordStore) Remove(file string) error {
	msg := fmt.Sprintf("Removed password file %q.", file)
	cb := func() error {
		return ps.store.RemoveFile(file)
	}
	if err := ps.store.Apply(msg, cb); err != nil {
		return xerrors.Errorf("could not remove password file %q: %w", file, err)
	}
	return nil
}

func (ps *PasswordStore) RemoveAll(file string) error {
	return xerrors.New("TODO")
}

func (ps *PasswordStore) Rename(oldpath, newpath string) error {
	msg := fmt.Sprintf("Renamed %q to %q.", oldpath, newpath)
	cb := func() error {
		return ps.store.Rename(oldpath, newpath)
	}
	if err := ps.store.Apply(msg, cb); err != nil {
		return xerrors.Errorf("could not rename password file %q to %q: %w", oldpath, newpath, err)
	}
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

// Format prepares password file content from a password and user data in the
// unencrypted form.
func Format(password string, data []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString(password + "\n")
	buf.Write(data)
	return buf.Bytes()
}

// Parse splits the decrypted password file content into a password and
// additional user data.
func Parse(decrypted []byte) (string, []byte) {
	index := bytes.IndexRune(decrypted, '\n')
	if index == -1 {
		return string(decrypted), nil
	}
	return string(decrypted[:index]), decrypted[index+1:]
}
