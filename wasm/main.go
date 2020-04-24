// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"log"
	"strings"

	"golang.org/x/xerrors"
	"honnef.co/go/js/dom/v2"
)

func main() {
	if err := doMain(); err != nil {
		log.Fatal(err)
	}
}

func doMain() error {
	window := dom.GetWindow()
	location := window.Location()
	protocol := location.Protocol()

	var backend Backend
	if strings.EqualFold(protocol, "chrome-extension:") {
		log.Println("Using chrome extension backend")
		b, err := NewExtensionBackend()
		if err != nil {
			return xerrors.Errorf("could not create chrome-extension backend: %w", err)
		}
		backend = b
	} else {
		log.Println("Using server backend with ajax")
		b, err := NewServerBackend(location.Href())
		if err != nil {
			return xerrors.Errorf("could not create server backend: %w", err)
		}
		backend = b
	}

	log.Println("Location", location)
	log.Println("Location.Protocol", location.Protocol())
	log.Println("Location.Host", location.Host())
	log.Println("Location.Hostname", location.Hostname())
	log.Println("Location.Href", location.Href())
	log.Println("Location.Origin", location.Origin())

	ctl, err := NewController(backend)
	if err != nil {
		return xerrors.Errorf("could not create controller: %w", err)
	}

	resp, err := ctl.CheckStatus()
	if err != nil {
		return xerrors.Errorf("could not check status: %w", err)
	}

	page, err := NewSettingsPage(ctl, resp)
	if err != nil {
		return xerrors.Errorf("could not create settings page: %w", err)
	}
	ctl.ShowPage(page)

	<-make(chan struct{})
	return nil
}
