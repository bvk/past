// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

const ExtensionID = "cbmmmlplcmfkgfbnhhiefheibgnpngff"

var installCmd = &cobra.Command{
	Use:   "install [flags] <extension-file>",
	Short: "Installs password-store browser extension for all Google-Chrome users.",
	RunE:  cmdInstall,
}

func init() {
	flags := installCmd.Flags()
	flags.Bool("use-opt-google", false, "When true, extension preferences file will be created in /opt/google/chrome")
	flags.String("extension-id", ExtensionID, "Chrome extension id.")
	flags.String("extension-version", "0.2", "Chrome extension version.")
	flags.String("native-message-host-name", "github.bvk.past", "Chrome native messaging host name.")
}

func cmdInstall(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()

	// This operation requires root access.
	if uid := os.Getuid(); uid != 0 {
		return xerrors.Errorf("only root can perform this operation: %w", os.ErrInvalid)
	}

	if len(args) == 0 {
		return xerrors.Errorf("need a .crx extension path argument: %w", os.ErrInvalid)
	}
	if len(args) > 1 {
		return xerrors.Errorf("too many arguments: %w", os.ErrInvalid)
	}
	useOptGoogle, err := flags.GetBool("use-opt-google")
	if err != nil {
		return xerrors.Errorf("could not get --use-opt-google value: %w", err)
	}
	extVersion, err := flags.GetString("extension-version")
	if err != nil {
		return xerrors.Errorf("could not get --extension-version value: %w", err)
	}
	extID, err := flags.GetString("extension-id")
	if err != nil {
		return xerrors.Errorf("could not get --extension-id value: %w", err)
	}
	nativeMsgHost, err := flags.GetString("native-message-host-name")
	if err != nil {
		return xerrors.Errorf("could not get --native-message-host-name value: %w", err)
	}

	var extData []byte
	if _, err := os.Stat(args[0]); err != nil {
		return xerrors.Errorf("could not stat extension at %q: %w", args[0], err)
	}
	if strings.HasPrefix(args[0], "https://") {
		resp, err := http.Get(args[0])
		if err != nil {
			return xerrors.Errorf("could not http-get from %q: %w", args[0], err)
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return xerrors.Errorf("could not read the package from %q: %w", args[0], err)
		}
		extData = data
	} else {
		absPath, err := filepath.Abs(args[0])
		if err != nil {
			return xerrors.Errorf("could not determine absolute path for %q: %w", args[0], err)
		}
		data, err := ioutil.ReadFile(absPath)
		if err != nil {
			return xerrors.Errorf("could not read binary file %q: %w", absPath, err)
		}
		extData = data
	}

	extDir := "/usr/lib/past"
	if _, err := os.Stat(extDir); err != nil {
		if !os.IsNotExist(err) {
			return xerrors.Errorf("could not stat %q: %w", extDir, err)
		}
		if err := os.Mkdir(extDir, os.FileMode(0755)); err != nil {
			return xerrors.Errorf("could not create directory %q: %w", extDir, err)
		}
	}
	extPath := filepath.Join(extDir, "past.crx")
	if err := ioutil.WriteFile(extPath, extData, os.FileMode(0644)); err != nil {
		return xerrors.Errorf("could not save the extension to %q: %w", extPath, err)
	}

	prefsDir := "/usr/share/google-chrome/extensions"
	if useOptGoogle {
		prefsDir = "/opt/google/chrome/extensions"
	}
	if _, err := os.Stat(prefsDir); err != nil {
		if !os.IsNotExist(err) {
			return xerrors.Errorf("could not stat extension preferences directory %q: %w", prefsDir, err)
		}
		if err := os.MkdirAll(prefsDir, os.FileMode(0755)); err != nil {
			return xerrors.Errorf("could not create extension preferences directory %q: %w", prefsDir, err)
		}
	}

	type ExtensionConfig struct {
		ExternalCRX     string `json:"external_crx"`
		ExternalVersion string `json:"external_version"`
	}
	var extCfg = &ExtensionConfig{
		ExternalCRX:     extPath,
		ExternalVersion: extVersion,
	}
	prefsData, err := json.MarshalIndent(extCfg, "", "    ")
	if err != nil {
		return xerrors.Errorf("could not marshal extension config: %w", err)
	}
	prefsPath := filepath.Join(prefsDir, extID+".json")
	if err := ioutil.WriteFile(prefsPath, prefsData, os.FileMode(0644)); err != nil {
		return xerrors.Errorf("could not create extension preferences file: %w", err)
	}

	//
	// Also, create a native messaging host configuration.
	//

	binaryPath, err := findBinaryPath(os.Args[0])
	if err != nil {
		return xerrors.Errorf("could not locate binary path: %w", err)
	}
	// Make a copy of the binary with the 'past' name.
	binaryData, err := ioutil.ReadFile(binaryPath)
	if err != nil {
		return xerrors.Errorf("could not read binary file %q: %w", binaryPath, err)
	}
	copyPath := filepath.Join(extDir, "past")
	if err := ioutil.WriteFile(copyPath, binaryData, os.FileMode(0755)); err != nil {
		return xerrors.Errorf("could not make a copy of the binary to %q: %w", copyPath, err)
	}

	nativeMsgHostsDir := "/etc/opt/chrome/native-messaging-hosts"
	if _, err := os.Stat(nativeMsgHostsDir); err != nil {
		if !os.IsNotExist(err) {
			return xerrors.Errorf("could not stat directory %q: %w", nativeMsgHostsDir, err)
		}
		if err := os.MkdirAll(nativeMsgHostsDir, os.FileMode(0755)); err != nil {
			return xerrors.Errorf("could not create native messaging host config directory %q: %w", nativeMsgHostsDir, err)
		}
	}

	type NativeMsgHostConfig struct {
		Name           string   `json:"name"`
		Description    string   `json:"description"`
		Path           string   `json:"path"`
		Type           string   `json:"type"`
		AllowedOrigins []string `json:"allowed_origins"`
	}
	var nativeCfg = &NativeMsgHostConfig{
		Name:        nativeMsgHost,
		Description: "Native messaging host config for past.",
		Path:        copyPath,
		Type:        "stdio",
		AllowedOrigins: []string{
			fmt.Sprintf("chrome-extension://%s/", extID),
		},
	}

	cfgData, err := json.MarshalIndent(nativeCfg, "", "    ")
	if err != nil {
		return xerrors.Errorf("could not marshal native messaging host config: %w", err)
	}

	file := filepath.Join(nativeMsgHostsDir, nativeCfg.Name+".json")
	if err := ioutil.WriteFile(file, cfgData, os.FileMode(0644)); err != nil {
		return xerrors.Errorf("could not write native messaging host config: %w", err)
	}
	return nil
}

func findBinaryPath(name string) (string, error) {
	bp, err := exec.LookPath(name)
	if err != nil {
		return "", xerrors.Errorf("could not determine current binary location: %w", err)
	}
	ap, err := filepath.Abs(bp)
	if err != nil {
		return "", xerrors.Errorf("could not determine absolute path for binary %q: %w", bp, err)
	}
	target, err := filepath.EvalSymlinks(ap)
	if err != nil {
		return "", xerrors.Errorf("could not locate current binary from %q: %w", ap, err)
	}
	return target, nil
}
