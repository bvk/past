// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"encoding/csv"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	ps, err := newPasswordStore(flags)
	if err != nil {
		return xerrors.Errorf("could not create password store instance: %w", err)
	}
	overwrite, err := flags.GetBool("overwrite")
	if err != nil {
		return xerrors.Errorf("could not get --overwrite value: %w", err)
	}
	ignoreFailures, err := flags.GetBool("ignore-failures")
	if err != nil {
		return xerrors.Errorf("could not get --ignore-failures value: %w", err)
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
	sitenameOffset := -1
	usernameOffset := -1
	passwordOffset := -1
	for i, header := range headers {
		if strings.EqualFold(header, "password") {
			passwordOffset = i
		}
		if strings.EqualFold(header, "name") {
			sitenameOffset = i
		}
		if strings.EqualFold(header, "username") {
			usernameOffset = i
		}
	}
	if passwordOffset == -1 {
		return xerrors.Errorf("chrome passwords file has no password column: %w", os.ErrInvalid)
	}
	if sitenameOffset == -1 {
		return xerrors.Errorf("chrome passwords file has no name column: %w", os.ErrInvalid)
	}
	if usernameOffset == -1 {
		return xerrors.Errorf("chrome passwords file has no username column: %w", os.ErrInvalid)
	}

	for i := 1; i < len(records); i++ {
		fields := records[i]
		if l := len(fields); l != len(headers) {
			log.Printf("line %d has not enough columns (ignored)", i)
			continue
		}

		vs := store.NewValues(nil)
		sitename := fields[sitenameOffset]
		username := fields[usernameOffset]
		password := fields[passwordOffset]
		for i, header := range headers {
			if i != sitenameOffset && i != passwordOffset && i != usernameOffset {
				vs.Set(header, fields[i])
			}
		}

		data := store.Format(password, vs.Bytes())
		filename := filepath.Join(sitename, username)

		if err := ps.CreateFile(filename, data, os.FileMode(0644)); err != nil {
			if overwrite && xerrors.Is(err, os.ErrExist) {
				err = ps.UpdateFile(filename, data)
			}
			if err != nil && !ignoreFailures {
				return xerrors.Errorf("could not add entry %s: %w", filename, err)
			}
			if err != nil {
				log.Printf("could not create or update password-file %q: %v", filename, err)
				continue
			}
		}
		log.Printf("added new password-file %q", filename)
	}
	return nil
}
