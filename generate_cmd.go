// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bvk/past/git"
	"github.com/bvk/past/gpg"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

var Symbols = `!"#$%&'()*+,\-./:;<=>?@[]^_{|}~` + "`"
var NonSymbols = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "abcdefghijklmnopqrstuvwxyz" + "0123456789"

var generateCmd = &cobra.Command{
	Use:   "generate [flags] <password-file>",
	Short: "Inserts a new password-file with an auto-generated password.",
	RunE:  cmdGenerate,
}

func init() {
	flags := generateCmd.Flags()
	flags.Bool("no-symbols", false, "When true, generated password contains only alphanumerics.")
	flags.Uint("length", 10, "Number of characters in the generated password.")
	flags.String("symbols", Symbols, "Acceptable set of symbols to use in the password.")
	flags.String("user", "", "Username to save along with the password.")
}

func cmdGenerate(cmd *cobra.Command, args []string) (status error) {
	flags := cmd.Flags()
	if len(args) == 0 {
		return xerrors.Errorf("search string argument is required: %w", os.ErrInvalid)
	}
	if len(args) > 1 {
		return xerrors.Errorf("too many search string arguments: %w", os.ErrInvalid)
	}
	file := filepath.Join("./", args[0]+".gpg")
	dataDir, err := flags.GetString("data-dir")
	if err != nil {
		return xerrors.Errorf("could not get --data-dir value: %w", err)
	}
	if len(dataDir) == 0 {
		return xerrors.Errorf("data directory path be empty: %w", os.ErrInvalid)
	}
	noSymbols, err := flags.GetBool("no-symbols")
	if err != nil {
		return xerrors.Errorf("could not get --no-symbols value: %w", err)
	}
	length, err := flags.GetUint("length")
	if err != nil {
		return xerrors.Errorf("could not get --length value: %w", err)
	}
	if length == 0 {
		return xerrors.Errorf("length value cannot be zero: %w", os.ErrInvalid)
	}
	user, err := flags.GetString("user")
	if err != nil {
		return xerrors.Errorf("could not get --user value: %w", err)
	}
	store, err := git.NewDir(dataDir)
	if err != nil {
		return xerrors.Errorf("could not create git directory instance: %w", err)
	}
	files, err := store.ListFiles()
	if err != nil {
		return xerrors.Errorf("could not list files in the git directory: %w", err)
	}
	// Scan all git files to identify .gpg-id files for each directory.
	keyDirs := make(map[string]struct{})
	for _, path := range files {
		base := filepath.Base(path)
		if base == ".gpg-id" {
			keyDirs[filepath.Dir(path)] = struct{}{}
		}
	}
	if _, ok := keyDirs["."]; !ok {
		return xerrors.Errorf("could not find the default gpg keys: %w", os.ErrInvalid)
	}

	keyring, err := gpg.NewKeyring("")
	if err != nil {
		return xerrors.Errorf("could not create gpg key ring instance: %w", err)
	}

	keys := []string{}
	for d := filepath.Dir(file); d != "."; d = filepath.Dir(d) {
		if _, ok := keyDirs[d]; ok {
			path := filepath.Join(d, ".gpg-id")
			data, err := store.ReadFile(path)
			if err != nil {
				return xerrors.Errorf("could not read file %q: %w", path, err)
			}
			keys = strings.Fields(string(data))
			break
		}
	}
	if len(keys) == 0 {
		file := ".gpg-id"
		data, err := store.ReadFile(file)
		if err != nil {
			return xerrors.Errorf("could not read file %q: %w", file, err)
		}
		keys = strings.Fields(string(data))
	}

	var pkeys []*gpg.PublicKey
	for _, pk := range keyring.PublicKeys() {
		for _, key := range keys {
			if pk.Fingerprint == key {
				pkeys = append(pkeys, pk)
			}
		}
	}
	if len(pkeys) == 0 {
		return xerrors.Errorf("gpg keyring doesn't have the required keys %q: %w", keys, os.ErrInvalid)
	}

	// Generate the password.
	var alnum = []byte(NonSymbols)
	var graph = []byte(NonSymbols + Symbols)
	sort.Slice(alnum, func(i, j int) bool { return alnum[i] < alnum[j] })
	sort.Slice(graph, func(i, j int) bool { return graph[i] < graph[j] })

	var buf bytes.Buffer
	if noSymbols {
		max := big.NewInt(int64(len(alnum)))
		for ii := uint(0); ii < length; ii++ {
			v, err := rand.Int(rand.Reader, max)
			if err != nil {
				return xerrors.Errorf("could not generate secure random int: %w", err)
			}
			buf.WriteByte(alnum[int(v.Int64())])
		}
	} else {
		max := big.NewInt(int64(len(graph)))
		for ii := uint(0); ii < length; ii++ {
			v, err := rand.Int(rand.Reader, max)
			if err != nil {
				return xerrors.Errorf("could not generate secure random int: %w", err)
			}
			buf.WriteByte(graph[int(v.Int64())])
		}
	}

	lines := []string{buf.String()}
	if len(user) > 0 {
		lines = append(lines, fmt.Sprintf("user: %s", user))
	}

	decrypted := []byte(strings.Join(lines, "\n") + "\n")
	encrypted, err := keyring.Encrypt(decrypted, pkeys)
	if err != nil {
		return xerrors.Errorf("could not encrypt new password: %w", err)
	}
	if err := store.AddFile(file, encrypted, os.FileMode(0644)); err != nil {
		return xerrors.Errorf("could not add file %q in git repo: %w", file, err)
	}
	defer func() {
		if status != nil {
			if err := store.Reset("HEAD"); err != nil {
				log.Panicf("could not undo adding file %q: %v", file, err)
				return
			}
		}
	}()

	msg := fmt.Sprintf("Added new password file %q", file)
	if err := store.Commit(msg); err != nil {
		return xerrors.Errorf("could not commit reinitialize change: %w", err)
	}
	return nil
}
