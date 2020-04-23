// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"os"
	"path/filepath"
	"sort"

	"github.com/bvk/past"

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
	ps, err := newPasswordStore(flags)
	if err != nil {
		return xerrors.Errorf("could not create password store instance: %w", err)
	}

	if len(args) == 0 {
		return xerrors.Errorf("search string argument is required: %w", os.ErrInvalid)
	}
	if len(args) > 1 {
		return xerrors.Errorf("too many search string arguments: %w", os.ErrInvalid)
	}
	file := filepath.Join("./", args[0])

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

	vs := past.NewValues(nil)
	if len(user) > 0 {
		vs.Set("username", user)
	}

	data := past.Format(buf.String(), vs.Bytes())
	if err := ps.CreateFile(file, data, os.FileMode(0644)); err != nil {
		return xerrors.Errorf("could not insert new file %q: %w", file, err)
	}
	return nil
}
