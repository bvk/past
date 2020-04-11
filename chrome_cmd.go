// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bvk/past/git"
	"github.com/bvk/past/gpg"
	"github.com/spf13/pflag"

	"golang.org/x/xerrors"
)

func cmdChrome(flags *pflag.FlagSet, args []string) error {
	// When invoked by chrome, PATH may not be the same, so fix it to a known
	// good defaults. For example, gpg command from GPGTools package in Mac OS X
	// is installed into /usr/local/bin directory, but it is not part of the PATH
	// when chrome extension invokes this program.
	homeBin := filepath.Join(os.Getenv("HOME"), "bin")
	os.Setenv("PATH", homeBin+":/bin:/usr/bin:/usr/local/bin:/sbin:/usr/sbin:/usr/local/sbin")

	dataDir, err := flags.GetString("data-dir")
	if err != nil {
		return xerrors.Errorf("could not get --data-dir value: %w", err)
	}
	if len(dataDir) == 0 {
		return xerrors.Errorf("data directory path be empty: %w", os.ErrInvalid)
	}
	store, err := git.NewDir(dataDir)
	if err != nil {
		return xerrors.Errorf("could not create git directory instance: %w", err)
	}
	keyring, err := gpg.NewKeyring("")
	if err != nil {
		return xerrors.Errorf("could not create gpg key ring instance: %w", err)
	}
	pstore, err := NewPasswordStore(store, keyring)
	if err != nil {
		return xerrors.Errorf("could not create password-store instance: %w", err)
	}
	h := ChromeHandler{
		store:   store,
		keyring: keyring,
		pstore:  pstore,
	}
	return h.ServeChrome(context.Background(), os.Stdin, os.Stdout)
}

type ChromeRequest struct {
	AddFile    *AddFileRequest    `json:"add_file"`
	EditFile   *EditFileRequest   `json:"edit_file"`
	ListFiles  *ListFilesRequest  `json:"list_files"`
	ViewFile   *ViewFileRequest   `json:"view_file"`
	DeleteFile *DeleteFileRequest `json:"delete_file"`
}

type ChromeResponse struct {
	// Status contains any error in performing the operation. It should be empty
	// on success.
	Status string `json:"status"`

	AddFile    *AddFileResponse    `json:"add_file"`
	EditFile   *EditFileResponse   `json:"edit_file"`
	ListFiles  *ListFilesResponse  `json:"list_files"`
	ViewFile   *ViewFileResponse   `json:"view_file"`
	DeleteFile *DeleteFileResponse `json:"delete_file"`
}

type ListFilesRequest struct {
}

type ListFilesResponse struct {
	Files []string `json:"files"`
}

type AddFileRequest struct {
	File     string      `json:"file"`
	Password string      `json:"password"`
	Username string      `json:"username"`
	Sitename string      `json:"sitename"`
	Rest     [][2]string `json:"rest"`
}

type AddFileResponse struct {
}

type EditFileRequest struct {
	File     string `json:"file"`
	OrigFile string `json:"orig_file"`

	Password string `json:"password"`
	Username string `json:"username"`
	Sitename string `json:"sitename"`
	Data     string `json:"data"`
}

type EditFileResponse struct {
}

type ViewFileRequest struct {
	File string `json:"file"`
}

type ViewFileResponse struct {
	Username      string      `json:"username"`
	Password      string      `json:"password"`
	KeyValuePairs [][2]string `json:"key_value_pairs"`
}

type DeleteFileRequest struct {
	File string `json:"file"`
}

type DeleteFileResponse struct {
}

type ChromeHandler struct {
	store   *git.Dir
	keyring *gpg.Keyring
	pstore  *PasswordStore
}

func (c *ChromeHandler) ServeChrome(ctx context.Context, in io.Reader, out io.Writer) error {
	var sizeBytes [4]byte
	if _, err := io.ReadFull(os.Stdin, sizeBytes[:]); err != nil {
		return xerrors.Errorf("could not read input message length: %w", err)
	}
	size := binary.LittleEndian.Uint32(sizeBytes[:])
	reqBuf := make([]byte, size)
	if _, err := io.ReadFull(os.Stdin, reqBuf); err != nil {
		return xerrors.Errorf("could not read input message: %w", err)
	}

	req := new(ChromeRequest)
	if err := json.Unmarshal(reqBuf, req); err != nil {
		return xerrors.Errorf("could not unmarshal input message: %w", err)
	}
	log.Printf("request -> %q", reqBuf)

	var resp ChromeResponse
	switch {
	case req.AddFile != nil:
		resp.AddFile = new(AddFileResponse)
		if err := c.doAddFile(ctx, req.AddFile, resp.AddFile); err != nil {
			resp.Status = err.Error()
		}
	case req.EditFile != nil:
		resp.EditFile = new(EditFileResponse)
		if err := c.doEditFile(ctx, req.EditFile, resp.EditFile); err != nil {
			resp.Status = err.Error()
		}
	case req.ListFiles != nil:
		resp.ListFiles = new(ListFilesResponse)
		if err := c.doListFiles(ctx, req.ListFiles, resp.ListFiles); err != nil {
			resp.Status = err.Error()
		}
	case req.ViewFile != nil:
		resp.ViewFile = new(ViewFileResponse)
		if err := c.doViewFile(ctx, req.ViewFile, resp.ViewFile); err != nil {
			resp.Status = err.Error()
		}
	case req.DeleteFile != nil:
		resp.DeleteFile = new(DeleteFileResponse)
		if err := c.doDeleteFile(ctx, req.DeleteFile, resp.DeleteFile); err != nil {
			resp.Status = err.Error()
		}
	default:
		resp.Status = xerrors.Errorf("unknown or invalid request: %w", os.ErrInvalid).Error()
	}

	respBytes, err := json.Marshal(&resp)
	if err != nil {
		return xerrors.Errorf("could not marshal response (%T) to json: %w", resp, err)
	}
	if err := binary.Write(os.Stdout, binary.LittleEndian, uint32(len(respBytes))); err != nil {
		return xerrors.Errorf("could not write response size: %w", err)
	}
	if _, err := os.Stdout.Write(respBytes); err != nil {
		return xerrors.Errorf("could not write response bytes: %w", err)
	}
	return nil
}

func (c *ChromeHandler) doAddFile(ctx context.Context, req *AddFileRequest, resp *AddFileResponse) error {
	var rest = [][2]string{
		[2]string{"user", strings.TrimSpace(req.Username)},
		[2]string{"site", strings.TrimSpace(req.Sitename)},
	}
	for _, other := range req.Rest {
		rest = append(rest, other)
	}

	if strings.Contains(req.File, "/") {
		return xerrors.Errorf("directories are not allowed in the file name: %w", os.ErrInvalid)
	}

	if err := c.pstore.AddPasswordFile(strings.TrimSpace(req.File), req.Password, rest); err != nil {
		return xerrors.Errorf("could not add new file: %w", err)
	}
	return nil
}

func (c *ChromeHandler) doEditFile(ctx context.Context, req *EditFileRequest, resp *EditFileResponse) error {
	var rest = [][2]string{
		[2]string{"user", strings.TrimSpace(req.Username)},
		[2]string{"site", strings.TrimSpace(req.Sitename)},
	}
	lines := strings.Split(req.Data, "\n")
	for _, line := range lines {
		line := strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if i := strings.IndexRune(line, ':'); i != -1 {
			key := strings.TrimSpace(line[:i])
			value := strings.TrimSpace(line[i+1:])
			if len(key) > 0 && len(value) > 0 {
				rest = append(rest, [2]string{key, value})
			}
			continue
		}
		rest = append(rest, [2]string{strings.TrimSpace(line), ""})
	}

	if strings.Contains(req.File, "/") {
		return xerrors.Errorf("directories are not allowed in the file name: %w", os.ErrInvalid)
	}

	file := strings.TrimSpace(req.File)
	if len(req.OrigFile) > 0 && req.OrigFile != req.File {
		origFile := strings.TrimSpace(req.OrigFile)
		if err := c.pstore.ReplacePasswordFile(origFile, file, req.Password, rest); err != nil {
			return xerrors.Errorf("could not replace file %q: %w", origFile, err)
		}
	} else {
		if err := c.pstore.UpdatePasswordFile(file, req.Password, rest); err != nil {
			return xerrors.Errorf("could not update file %q: %w", req.File, err)
		}
	}
	return nil
}

func (c *ChromeHandler) doListFiles(ctx context.Context, req *ListFilesRequest, resp *ListFilesResponse) error {
	files, err := c.store.ListFiles()
	if err != nil {
		return xerrors.Errorf("could not list files in the git directory: %w", err)
	}

	for _, file := range files {
		// We don't want to support directory structure with the extension.
		if strings.ContainsRune(file, filepath.Separator) {
			continue
		}
		if strings.HasSuffix(file, ".gpg") {
			resp.Files = append(resp.Files, strings.TrimSuffix(file, ".gpg"))
		}
	}
	return nil
}

func (c *ChromeHandler) doViewFile(ctx context.Context, req *ViewFileRequest, resp *ViewFileResponse) error {
	file := filepath.Join("./", req.File+".gpg")
	encrypted, err := c.store.ReadFile(file)
	if err != nil {
		return xerrors.Errorf("could not get login entry with name %q: %w", req.File, err)
	}
	decrypted, err := c.keyring.Decrypt(encrypted)
	if err != nil {
		return xerrors.Errorf("could not decrypt login entry %q: %w", req.File, err)
	}
	// Identify the username here, so that UI layer becomes easier.
	password, kvs := parse(decrypted)
	for _, kv := range kvs {
		key := strings.ToLower(kv[0])
		if len(resp.Username) == 0 && key == "username" || key == "user" || key == "login" {
			resp.Username = kv[1]
		} else {
			resp.KeyValuePairs = append(resp.KeyValuePairs, kv)
		}
	}
	resp.Password = password
	return nil
}

func (c *ChromeHandler) doDeleteFile(ctx context.Context, req *DeleteFileRequest, resp *DeleteFileResponse) error {
	file := filepath.Clean(filepath.Join("./", req.File+".gpg"))
	if err := c.store.RemoveFile(file); err != nil {
		return xerrors.Errorf("could not remove file %q: %w", file, err)
	}
	return nil
}

func parse(data []byte) (string, [][2]string) {
	lines := bytes.Split(data, []byte("\n"))
	values := [][2]string{}
	if len(lines) == 0 {
		return "", values
	}
	for ii := 1; ii < len(lines); ii++ {
		s := string(lines[ii])
		if colon := strings.IndexRune(s, ':'); colon >= 0 {
			key := strings.TrimSpace(s[:colon])
			value := strings.TrimSpace(s[colon+1:])
			values = append(values, [2]string{key, value})
		}
	}
	password := strings.TrimSpace(string(lines[0]))
	return password, values
}
