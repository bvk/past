// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"strings"
	"syscall/js"

	"github.com/bvk/past/msg"
	"golang.org/x/xerrors"
)

type StorageSupport interface {
	GetItem(key string) (string, bool)
	SetItem(key, val string)
}

type ContextMenuSupport interface {
	RemoveAllContextMenus()
	RemoveContextMenu(id string) error
	AddContextMenu(id, title string, menuContexts, urlPatterns []string) error

	// NOTE: Context menu handler is configured in the background page cause it
	// needs to run when popup window is not active.
}

type ExtensionBackend struct {
	tabAddr string

	bg js.Value
}

var _ ContextMenuSupport = &ExtensionBackend{}

func NewExtensionBackend() (*ExtensionBackend, error) {
	e := &ExtensionBackend{}
	tab := ActiveTab()
	// FIXME: We should use js.Value types' Equal method with go1.13+
	if tab != js.Undefined() && tab != js.Null() {
		url := tab.Get("url")
		e.tabAddr = url.String()
	}
	bg := GetBackgroundPage()
	if bg == js.Undefined() || bg == js.Null() {
		return nil, xerrors.Errorf("could not get background page reference: %w", os.ErrInvalid)
	}
	e.bg = bg
	return e, nil
}

func (e *ExtensionBackend) Call(req *msg.Request) (*msg.Response, error) {
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
	resp := new(msg.Response)
	if err := js2go(result, resp); err != nil {
		return nil, xerrors.Errorf("could not decode the native messaging host response: %w", err)
	}
	return resp, nil
}

func (e *ExtensionBackend) ActiveTabAddr() string {
	return e.tabAddr
}

func (e *ExtensionBackend) AddContextMenu(id, name string, menuContexts, urlMatches []string) error {
	r := AddContextMenu(id, name, menuContexts, urlMatches)
	if r == js.Undefined() || r == js.Null() {
		return xerrors.Errorf("could not create context menu entry: %w", os.ErrInvalid)
	}
	return nil
}

func (e *ExtensionBackend) RemoveContextMenu(id string) error {
	RemoveContextMenu(id)
	return nil
}

func (e *ExtensionBackend) RemoveAllContextMenus() {
	RemoveAllContextMenus()
}

func (e *ExtensionBackend) GetItem(key string) (string, bool) {
	v := GetItem(key)
	if v == js.Undefined() || v == js.Null() {
		return "", false
	}
	m := make(map[string]string)
	if err := js2go(v, m); err != nil {
		log.Printf("error: could not convert %q to %T", js2str(v), m)
		return "", false
	}
	if val, ok := m[key]; ok {
		return val, true
	}
	return "", false
}

func (e *ExtensionBackend) SetItem(key, val string) {
	SetItem(key, val)
}

func js2go(src js.Value, dst interface{}) error {
	result := js.Global().Get("JSON").Call("stringify", src)
	reader := strings.NewReader(result.String())
	return json.NewDecoder(reader).Decode(dst)
}

func js2str(src js.Value) string {
	result := js.Global().Get("JSON").Call("stringify", src)
	return result.String()
}
