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
	h := ChromeHandler{
		store:   store,
		keyring: keyring,
	}
	return h.ServeChrome(context.Background(), os.Stdin, os.Stdout)
}

type ChromeRequest struct {
	ListFiles *ListFilesRequest `json:"list_files"`
	ViewFile  *ViewFileRequest  `json:"view_file"`
}

type ChromeResponse struct {
	// Status contains any error in performing the operation. It should be empty
	// on success.
	Status string `json:"status"`

	ListFiles *ListFilesResponse `json:"list_files"`
	ViewFile  *ViewFileResponse  `json:"view_file"`
}

type ListFilesRequest struct {
}

type ListFilesResponse struct {
	Files []string `json:"files"`
}

type ViewFileRequest struct {
	File string `json:"file"`
}

type ViewFileResponse struct {
	Password string      `json:"password"`
	Values   [][2]string `json:"values"`
}

type ChromeHandler struct {
	store   *git.Dir
	keyring *gpg.Keyring
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
	log.Printf("response -> %q", respBytes)
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
	resp.Password, resp.Values = parse(decrypted)
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
			values = append(values, [2]string{s[:colon], s[colon+1:]})
		}
	}
	password := strings.TrimSpace(string(lines[0]))
	return password, values
}
