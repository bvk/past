// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/bvk/past/gpg"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

var ExtensionIDs = []string{
	"lpjgobmcekjengejhfbambleokkelpjb", // Chrome Web Store version.
	"klnbabongmhnpiffdkadmanmcfcneacn", // Local developement version.
}

// Mappings for native-messaging-hosts config file directory location (for
// current-user and system-wide) on each os for each browser.
var nativeMsgHostsDirMap = map[string]map[string][2]string{
	"linux": map[string][2]string{
		"google-chrome": [2]string{
			filepath.Join(os.Getenv("HOME"), ".config/google-chrome/NativeMessagingHosts"),
			"/etc/opt/chrome/native-messaging-hosts",
		},

		"chromium": [2]string{
			filepath.Join(os.Getenv("HOME"), ".config/chromium/NativeMessagingHosts"),
			"/etc/chromium/native-messaging-hosts/",
		},
	},
	"darwin": map[string][2]string{
		"google-chrome": [2]string{
			filepath.Join(os.Getenv("HOME"), "/Library/Application Support/Google/Chrome/NativeMessagingHosts/"),
			"/Library/Google/Chrome/NativeMessagingHosts/",
		},
		"chromium": [2]string{
			filepath.Join(os.Getenv("HOME"), "/Library/Application Support/Chromium/NativeMessagingHosts/"),
			"/Library/Application Support/Chromium/NativeMessagingHosts/",
		},
	},
}

var installCmd = &cobra.Command{
	Use:   "install [flags]",
	Short: "Installs the backend for browser extension.",
	RunE:  cmdInstall,
}

func init() {
	flags := installCmd.Flags()
	flags.Bool("dev", false, "When true, backend will be configured for extension development.")
	flags.String("browser", "google-chrome", "Command name to open the browser.")
	flags.String("extension-dir", "", "Directory where past binary will be copied.")
	flags.String("native-message-hosts-dir", "", "Directory where native messaging host config is created.")
	flags.String("native-message-hosts-name", "github.bvk.past", "Chrome native messaging host name.")
}

func cmdInstall(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()

	uid := os.Getuid()
	if uid != 0 {
		keyring, err := gpg.NewKeyring("")
		if err != nil {
			return xerrors.Errorf("could not scan for gpg keyring: %w", err)
		}
		pks := keyring.PublicKeys()
		if len(pks) == 0 {
			log.Printf("warning: gpg keyring has no keys; browser extension may not be useful")
		}

		dataDir, err := flags.GetString("data-dir")
		if err != nil {
			return xerrors.Errorf("could not get --data-dir value: %w", err)
		}
		if _, err := os.Stat(dataDir); err != nil {
			if !os.IsNotExist(err) {
				return xerrors.Errorf("could not stat password-store at %q: %w", dataDir, err)
			}
			log.Printf("password-store directory %q doesn't exist; browser extension may not be useful")
		}
	}

	browser, err := flags.GetString("browser")
	if err != nil {
		return xerrors.Errorf("could not get --browser value: %w", err)
	}

	dev, err := flags.GetBool("dev")
	if err != nil {
		return xerrors.Errorf("could not get --dev value: %w", err)
	}
	extID := ExtensionIDs[0]
	if dev {
		extID = ExtensionIDs[1]
	}
	extDir, err := flags.GetString("extension-dir")
	if err != nil {
		return xerrors.Errorf("could not get --extension-dir value: %w", err)
	}
	nativeMsgHostsName, err := flags.GetString("native-message-hosts-name")
	if err != nil {
		return xerrors.Errorf("could not get --native-message-hosts-name value: %w", err)
	}
	nativeMsgHostsDir, err := flags.GetString("native-message-hosts-dir")
	if err != nil {
		return xerrors.Errorf("could not get --native-message-hosts-dir value: %w", err)
	}

	if len(extDir) == 0 {
		if uid == 0 {
			extDir = "/usr/local/bin"
		} else {
			extDir = filepath.Join(os.Getenv("HOME"), ".config/past")
		}
	}

	// Find the appropriate directory for native messaging host config files.
	if len(nativeMsgHostsDir) == 0 {
		xx, ok := nativeMsgHostsDirMap[runtime.GOOS]
		if !ok {
			return xerrors.Errorf("os %q is not supported: %w", runtime.GOOS, os.ErrInvalid)
		}
		yy, ok := xx[browser]
		if !ok {
			return xerrors.Errorf("browser %q is not supported: %w", browser, os.ErrInvalid)
		}
		nativeMsgHostsDir = yy[0]
		if uid == 0 {
			nativeMsgHostsDir = yy[1]
		}
	}

	// Make a copy of this binary in the extension directory.
	if _, err := os.Stat(extDir); err != nil {
		if !os.IsNotExist(err) {
			return xerrors.Errorf("could not stat %q: %w", extDir, err)
		}
		if err := os.Mkdir(extDir, os.FileMode(0755)); err != nil {
			return xerrors.Errorf("could not create directory %q: %w", extDir, err)
		}
	}
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
	log.Printf("past binary is copied to %q", copyPath)

	// Create a native messaging hosts config file.
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
		Name:        nativeMsgHostsName,
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
	log.Printf("native messaging host config file is created at %q", file)

	// Open chrome to prompt the user for installing the extension.
	if extID == ExtensionIDs[0] {
		address := "https://chrome.google.com/webstore/detail/password-store-extension/lpjgobmcekjengejhfbambleokkelpjb"
		if err := openBrowser(browser, address); err != nil {
			log.Printf("visit %q in your browser to install the extension manually")
		}
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

func openBrowser(browser, address string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command(browser, address).Run()
	case "darwin":
		if browser == "google-chrome" {
			return exec.Command("open", "-a", "Google Chrome", address).Run()
		}
		if browser == "chromium" {
			return exec.Command("open", "-a", "Chromium", address).Run()
		}
	}
	return os.ErrInvalid
}
