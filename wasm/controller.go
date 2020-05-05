// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"strings"
	"syscall/js"
	"time"

	"golang.org/x/xerrors"
	"honnef.co/go/js/dom/v2"

	"github.com/bvk/past/msg"
	storage "github.com/bvk/past/wasm/denwc-storage"
)

type Page interface {
	RootDiv() dom.Element
	RefreshDisplay() error
}

type Backend interface {
	ActiveTabAddr() string
	Call(*msg.Request) (*msg.Response, error)
}

type Browser interface {
	Copy(string) error
	NativeCall(req interface{}) (interface{}, error)
}

type State struct {
	UseCount map[string]int
}

type Controller struct {
	backend Backend

	window dom.Window

	document dom.HTMLDocument

	body dom.HTMLElement

	currentPage Page

	state *State
}

func NewController(backend Backend) (*Controller, error) {
	w := dom.GetWindow()
	d := w.Document()
	hd, ok := d.(dom.HTMLDocument)
	if !ok {
		return nil, xerrors.Errorf("could not convert Document to HTMLDocument: %w", os.ErrInvalid)
	}
	body := hd.Body()
	if body == nil {
		return nil, xerrors.Errorf("could not find Body element: %w", os.ErrInvalid)
	}
	c := &Controller{
		window:   w,
		document: hd,
		body:     body,
		backend:  backend,
	}
	state, err := c.getState()
	if err != nil {
		return nil, xerrors.Errorf("could not read persistent state: %w", err)
	}
	c.state = state
	return c, nil
}

func (c *Controller) getState() (*State, error) {
	local := storage.Local()
	data, ok := local.GetItem("past")
	if !ok {
		log.Println("found no persistent state saved previously")
		return new(State), nil
	}
	state := new(State)
	if err := json.NewDecoder(strings.NewReader(data)).Decode(state); err != nil {
		return nil, err
	}
	log.Println("last persistent state", data)
	return state, nil
}

func (c *Controller) UpdateUseCount(useMap map[string]int) error {
	dupMap := make(map[string]int)
	for k, v := range useMap {
		dupMap[k] = v
	}
	c.state.UseCount = dupMap
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(c.state); err != nil {
		return xerrors.Errorf("could not encode persistent state: %w", err)
	}
	local := storage.Local()
	local.SetItem("past", b.String())
	return nil
}

func (c *Controller) ClosePage(p Page) error {
	w := dom.GetWindow()
	w.Close()
	return nil
}

func (c *Controller) ShowPage(page Page) error {
	if err := page.RefreshDisplay(); err != nil {
		return err
	}
	c.body.ReplaceChild(page.RootDiv(), c.body.FirstChild())
	c.currentPage = page
	return nil
}

func (c *Controller) UseCountMap() map[string]int {
	return c.state.UseCount
}

func (c *Controller) CheckStatus() (*msg.Response, error) {
	req := msg.Request{
		CheckStatus: new(msg.CheckStatusRequest),
	}
	return c.backend.Call(&req)
}

func (c *Controller) PasswordStoreStatus() (*msg.ScanStoreResponse, error) {
	req := msg.Request{
		ScanStore: new(msg.ScanStoreRequest),
	}
	resp, err := c.backend.Call(&req)
	if err != nil {
		return nil, err
	}
	return resp.ScanStore, nil
}

func (c *Controller) RemoteStatus() (*msg.SyncRemoteResponse, error) {
	req := msg.Request{
		SyncRemote: new(msg.SyncRemoteRequest),
	}
	resp, err := c.backend.Call(&req)
	if err != nil {
		return nil, err
	}
	return resp.SyncRemote, nil
}

func (c *Controller) SyncRemote() (*msg.SyncRemoteResponse, error) {
	req := msg.Request{
		SyncRemote: new(msg.SyncRemoteRequest),
	}
	resp, err := c.backend.Call(&req)
	if err != nil {
		return nil, err
	}
	return resp.SyncRemote, nil
}

func (c *Controller) PullRemote() (*msg.SyncRemoteResponse, error) {
	req := msg.Request{
		SyncRemote: &msg.SyncRemoteRequest{
			Pull: true,
		},
	}
	resp, err := c.backend.Call(&req)
	if err != nil {
		return nil, err
	}
	return resp.SyncRemote, nil
}

func (c *Controller) PushRemote() (*msg.SyncRemoteResponse, error) {
	req := msg.Request{
		SyncRemote: &msg.SyncRemoteRequest{
			Push: true,
		},
	}
	resp, err := c.backend.Call(&req)
	if err != nil {
		return nil, err
	}
	return resp.SyncRemote, nil
}

func (c *Controller) ImportKey(req *msg.ImportKeyRequest) (*msg.ImportKeyResponse, error) {
	resp, err := c.backend.Call(&msg.Request{ImportKey: req})
	if err != nil {
		return nil, err
	}
	return resp.ImportKey, nil
}

func (c *Controller) CreateKey(req *msg.CreateKeyRequest) (*msg.CreateKeyResponse, error) {
	resp, err := c.backend.Call(&msg.Request{CreateKey: req})
	if err != nil {
		return nil, err
	}
	return resp.CreateKey, nil
}

func (c *Controller) EditKey(req *msg.EditKeyRequest) (*msg.EditKeyResponse, error) {
	resp, err := c.backend.Call(&msg.Request{EditKey: req})
	if err != nil {
		return nil, err
	}
	return resp.EditKey, nil
}

func (c *Controller) ExportKey(req *msg.ExportKeyRequest) (*msg.ExportKeyResponse, error) {
	resp, err := c.backend.Call(&msg.Request{ExportKey: req})
	if err != nil {
		return nil, err
	}
	return resp.ExportKey, nil
}

func (c *Controller) DeleteKey(req *msg.DeleteKeyRequest) (*msg.DeleteKeyResponse, error) {
	resp, err := c.backend.Call(&msg.Request{DeleteKey: req})
	if err != nil {
		return nil, err
	}
	return resp.DeleteKey, nil
}

func (c *Controller) CreateRepo(req *msg.CreateRepoRequest) (*msg.CreateRepoResponse, error) {
	resp, err := c.backend.Call(&msg.Request{CreateRepo: req})
	if err != nil {
		return nil, err
	}
	return resp.CreateRepo, nil
}

func (c *Controller) ImportRepo(req *msg.ImportRepoRequest) (*msg.ImportRepoResponse, error) {
	resp, err := c.backend.Call(&msg.Request{ImportRepo: req})
	if err != nil {
		return nil, err
	}
	return resp.ImportRepo, nil
}

func (c *Controller) AddRecipient(req *msg.AddRecipientRequest) (*msg.AddRecipientResponse, error) {
	resp, err := c.backend.Call(&msg.Request{AddRecipient: req})
	if err != nil {
		return nil, err
	}
	return resp.AddRecipient, nil
}

func (c *Controller) RemoveRecipient(req *msg.RemoveRecipientRequest) (*msg.RemoveRecipientResponse, error) {
	resp, err := c.backend.Call(&msg.Request{RemoveRecipient: req})
	if err != nil {
		return nil, err
	}
	return resp.RemoveRecipient, nil
}

func (c *Controller) AddRemote(req *msg.AddRemoteRequest) (*msg.AddRemoteResponse, error) {
	resp, err := c.backend.Call(&msg.Request{AddRemote: req})
	if err != nil {
		return nil, err
	}
	return resp.AddRemote, nil
}

func (c *Controller) ListFiles() (*msg.ListFilesResponse, error) {
	resp, err := c.backend.Call(&msg.Request{ListFiles: new(msg.ListFilesRequest)})
	if err != nil {
		return nil, err
	}
	return resp.ListFiles, nil
}

func (c *Controller) AddFile(req *msg.AddFileRequest) (*msg.AddFileResponse, error) {
	resp, err := c.backend.Call(&msg.Request{AddFile: req})
	if err != nil {
		return nil, err
	}
	return resp.AddFile, nil
}

func (c *Controller) ViewFile(req *msg.ViewFileRequest) (*msg.ViewFileResponse, error) {
	resp, err := c.backend.Call(&msg.Request{ViewFile: req})
	if err != nil {
		return nil, err
	}
	return resp.ViewFile, nil
}

func (c *Controller) EditFile(req *msg.EditFileRequest) (*msg.EditFileResponse, error) {
	resp, err := c.backend.Call(&msg.Request{EditFile: req})
	if err != nil {
		return nil, err
	}
	return resp.EditFile, nil
}

func (c *Controller) DeleteFile(req *msg.DeleteFileRequest) (*msg.DeleteFileResponse, error) {
	resp, err := c.backend.Call(&msg.Request{DeleteFile: req})
	if err != nil {
		return nil, err
	}
	return resp.DeleteFile, nil
}

func (c *Controller) Copy(s string) error {
	return c.CopyTimeout(s, 0)
}

func (c *Controller) CopyTimeout(str string, timeout time.Duration) error {
	// See https://htmldom.dev/copy-text-to-the-clipboard

	// Create a fake textarea
	ta := c.document.CreateElement("textarea").(*dom.HTMLTextAreaElement)
	ta.SetValue(str)

	// Reset styles
	taStyle := ta.Style()
	taStyle.SetProperty("border", "0", "")
	taStyle.SetProperty("padding", "0", "")
	taStyle.SetProperty("margin", "0", "")

	// Set the absolute position
	// User won't see the element
	taStyle.SetProperty("position", "absolute", "")
	taStyle.SetProperty("left", "-9999px", "")
	taStyle.SetProperty("top", "0px", "")

	// Append the textarea to body
	c.body.AppendChild(ta)

	// Focus and select the text
	ta.Focus()
	ta.Select()

	d := js.ValueOf(c.document)
	d.Call("execCommand", "copy")
	c.body.RemoveChild(ta)

	// Clear the password after 10 seconds.
	if str != "*" {
		if timeout > 0 {
			go func() {
				<-time.After(10 * time.Second)
				c.CopyTimeout("*", 0)
			}()
		}
	}

	return nil
}

func (c *Controller) ActiveTabAddress() string {
	return ""
}

func (c *Controller) setGoodStatus(text string) {
	log.Println(text)
	c.setBadStatus("")
}

func (c *Controller) setError(err error) error {
	c.setBadStatus(err.Error())
	return err
}

func (c *Controller) setBadStatus(text string) {
	if len(text) > 0 {
		log.Printf("error: %s", text)
	}

	if c.currentPage != nil {
		root := c.currentPage.RootDiv()
		statuses := root.GetElementsByClassName("footer-status")
		for _, status := range statuses {
			status.SetTextContent(text)
		}
	}
}
