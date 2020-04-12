// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"encoding/csv"
	"log"
	"os"
	"strings"

	"github.com/bvk/past/git"
	"github.com/bvk/past/gpg"
	"github.com/bvk/past/store"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/xerrors"
)

var importCmd = &cobra.Command{
	Use:   "import <flag>...",
	Short: "Imports passwords from other password managers' data files.",
	RunE:  cmdImport,
}

func init() {
	flags := importCmd.Flags()
	flags.Bool("overwrite", false, "When true, existing files with matching name will be overwritten.")
	flags.Bool("ignore-failures", false, "When true, failures are ignored till all entries are processed.")
	flags.String("chrome-passwords-file", "", "Path to Chrome passwords data file.")
}

func cmdImport(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	chromePasswordsFile, err := flags.GetString("chrome-passwords-file")
	if err != nil {
		return xerrors.Errorf("could not get --chrome-passwords-file value: %w", err)
	}
	imported := false
	if len(chromePasswordsFile) > 0 {
		if err := importChromePasswords(flags, chromePasswordsFile); err != nil {
			return xerrors.Errorf("could not import chrome passwords from %q: %w", chromePasswordsFile, err)
		}
		imported = true
	}
	if !imported {
		return xerrors.Errorf("use one of the flags to specify password data file: %w", os.ErrInvalid)
	}
	return nil
}

func importChromePasswords(flags *pflag.FlagSet, file string) error {
	dataDir, err := flags.GetString("data-dir")
	if err != nil {
		return xerrors.Errorf("could not get --data-dir value: %w", err)
	}
	overwrite, err := flags.GetBool("overwrite")
	if err != nil {
		return xerrors.Errorf("could not get --overwrite value: %w", err)
	}
	ignoreFailures, err := flags.GetBool("ignore-failures")
	if err != nil {
		return xerrors.Errorf("could not get --ignore-failures value: %w", err)
	}
	repo, err := git.NewDir(dataDir)
	if err != nil {
		return xerrors.Errorf("could not create git directory instance: %w", err)
	}
	keyring, err := gpg.NewKeyring("")
	if err != nil {
		return xerrors.Errorf("could not create gpg key ring instance: %w", err)
	}
	ps, err := store.New(repo, keyring)
	if err != nil {
		return xerrors.Errorf("could not create password store instance: %w", err)
	}

	f, err := os.Open(file)
	if err != nil {
		return xerrors.Errorf("could not open chrome passwords file %q: %w", file, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return xerrors.Errorf("could not parse chrome passwords csv file: %w", err)
	}
	if len(records) == 0 {
		return xerrors.Errorf("chrome passwords file has no headers: %w", os.ErrInvalid)
	}

	headers := records[0]
	nameOffset := -1
	passwordOffset := -1
	for i, header := range headers {
		if strings.EqualFold(header, "password") {
			passwordOffset = i
		}
		if strings.EqualFold(header, "name") {
			nameOffset = i
		}
	}
	if passwordOffset == -1 {
		return xerrors.Errorf("chrome passwords file has no password column: %w", os.ErrInvalid)
	}
	if nameOffset == -1 {
		return xerrors.Errorf("chrome passwords file has no name column: %w", os.ErrInvalid)
	}

	for i := 1; i < len(records); i++ {
		fields := records[i]
		if l := len(fields); l != len(headers) {
			log.Printf("line %d has not enough columns (ignored)", i)
			continue
		}
		var rest [][2]string
		name, password := fields[nameOffset], fields[passwordOffset]
		for i, header := range headers {
			if i != nameOffset && i != passwordOffset {
				rest = append(rest, [2]string{header, fields[i]})
			}
		}
		if err := ps.AddPasswordFile(name, password, rest); err != nil {
			if overwrite && xerrors.Is(err, os.ErrExist) {
				err = ps.UpdatePasswordFile(name, password, rest)
			}
			if err != nil && !ignoreFailures {
				return xerrors.Errorf("could not add entry %s: %w", name, err)
			}
			if err != nil {
				log.Printf("could not create or update password-file %q: %v", name, err)
				continue
			}
		}
		log.Printf("added new password-file %q", name)
	}
	return nil
}
