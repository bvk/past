// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

var scanCmd = &cobra.Command{
	Use:   "scan [flags] <regexp|substr>",
	Short: "Decrypts all files to search for a string or regexp.",
	RunE:  cmdScan,
}

func init() {
	flags := scanCmd.Flags()
	flags.Bool("skip-decrypt-failures", false, "When true, files that could not be decrypted will be skipped.")
	flags.Bool("regexp", false, "When true, the search string argument is treated as a Go regexp.")
}

func cmdScan(cmd *cobra.Command, args []string) error {
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
	isRegexp, err := flags.GetBool("regexp")
	if err != nil {
		return xerrors.Errorf("could not get --regexp value: %w", err)
	}
	var re *regexp.Regexp
	if isRegexp {
		exp, err := regexp.Compile(regexp.QuoteMeta(args[0]))
		if err != nil {
			return xerrors.Errorf("search argument %q is not a valid regexp: %w", args[0], err)
		}
		re = exp
	}
	skipDecryptFailures, err := flags.GetBool("skip-decrypt-failures")
	if err != nil {
		return xerrors.Errorf("could not get --skip-decrypt-failures value: %w", err)
	}

	files, err := ps.ListFiles()
	if err != nil {
		return xerrors.Errorf("could not list files in the password store: %w", err)
	}

	skipped := []string{}
	for _, file := range files {
		decrypted, err := ps.ReadFile(file)
		if err != nil {
			if !skipDecryptFailures {
				return xerrors.Errorf("could not read file %q: %w", file, err)
			}
			skipped = append(skipped, file)
		}
		for ii, line := range bytes.Split(decrypted, []byte("\n")) {
			if re != nil {
				if v := re.Find(line); v != nil {
					fmt.Printf("%s:%d: %s\n", file, ii, line)
				}
				continue
			}
			if bytes.Contains(line, []byte(args[0])) {
				fmt.Printf("%s:%d: %s\n", file, ii, line)
			}
		}
	}

	if len(skipped) > 0 {
		log.Printf("warning: could not decrypt files %q, so they are skipped", skipped)
	}
	return nil
}
