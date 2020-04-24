// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"syscall/js"

	"github.com/bvk/past/msg"
	"golang.org/x/xerrors"
)

type ExtensionBackend struct {
	tabAddr string
}

func NewExtensionBackend() (*ExtensionBackend, error) {
	e := &ExtensionBackend{}
	tab := ActiveTab()
	// FIXME: We should use js.Value types' Equal method with go1.13+
	if tab != js.Undefined() && tab != js.Null() {
		url := tab.Get("url")
		e.tabAddr = url.String()
	}
	return e, nil
}

func (e *ExtensionBackend) Call(req *msg.BrowserRequest) (*msg.BrowserResponse, error) {
	var in bytes.Buffer
	if err := json.NewEncoder(&in).Encode(req); err != nil {
		return nil, err
	}
	obj := make(map[string]interface{})
	if err := json.Unmarshal(in.Bytes(), &obj); err != nil {
		return nil, err
	}

	result := SendNativeMessage("github.bvk.past", js.ValueOf(obj))
	// FIXME: We should use js.Value types' Equal method with go1.13+
	if result == js.Null() || result == js.Undefined() {
		return nil, xerrors.Errorf("could not call the native messaging host backend: %w", os.ErrInvalid)
	}
	resp := new(msg.BrowserResponse)
	if err := js2go(result, resp); err != nil {
		return nil, xerrors.Errorf("could not decode the native messaging host response: %w", err)
	}
	return resp, nil
}

func (e *ExtensionBackend) ActiveTabAddr() string {
	return e.tabAddr
}

func js2go(src js.Value, dst interface{}) error {
	result := js.Global().Get("JSON").Call("stringify", src)
	reader := strings.NewReader(result.String())
	return json.NewDecoder(reader).Decode(dst)
}
