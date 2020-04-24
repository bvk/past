// Copyright (c) 2020 BVK Chaitanya

package main

import "testing"

func TestParseTemplates(t *testing.T) {
	settingsPage := new(SettingsPage)
	if _, err := BindHTML(settingsTemplate, "gotag", settingsPage); err != nil {
		t.Errorf("could not parse settings page template: %v", err)
		return
	}

	addKeyPage := new(AddKeyPage)
	if _, err := BindHTML(addKeyTemplate, "gotag", addKeyPage); err != nil {
		t.Errorf("could not parse add key page template: %v", err)
		return
	}

	keyringPage := new(KeyringPage)
	if _, err := BindHTML(keyringTemplate, "gotag", keyringPage); err != nil {
		t.Errorf("could not parse keyring page template: %v", err)
		return
	}

	keyPage := new(KeyPage)
	if _, err := BindHTML(keyTemplate, "gotag", keyPage); err != nil {
		t.Errorf("could not parse key page template: %v", err)
		return
	}

	initPastPage := new(InitPastPage)
	if _, err := BindHTML(initPastTemplate, "gotag", initPastPage); err != nil {
		t.Errorf("could not parse init password store page template: %v", err)
		return
	}

	pastPage := new(PastPage)
	if _, err := BindHTML(pastTemplate, "gotag", pastPage); err != nil {
		t.Errorf("could not parse password store page template: %v", err)
		return
	}

	initRemotePage := new(InitRemotePage)
	if _, err := BindHTML(initRemoteTemplate, "gotag", initRemotePage); err != nil {
		t.Errorf("could not parse add init remote page template: %v", err)
		return
	}

	remotePage := new(RemotePage)
	if _, err := BindHTML(remoteTemplate, "gotag", remotePage); err != nil {
		t.Errorf("could not parse add remote page template: %v", err)
		return
	}
}
