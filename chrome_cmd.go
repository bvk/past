// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/bvk/past/git"
	"github.com/bvk/past/gpg"
	"github.com/bvk/past/store"

	"github.com/spf13/pflag"
	"golang.org/x/xerrors"
)

func cmdChrome(flags *pflag.FlagSet, args []string) (status error) {
	// Redirect the logs to a file.
	file := filepath.Join(os.TempDir(), fmt.Sprintf("past-%s.log", os.Getenv("USER")))
	logfile, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.FileMode(0600))
	if err != nil {
		return xerrors.Errorf("could not open log file %q: %w", file, err)
	}
	log.SetOutput(logfile)
	defer func() {
		if status != nil {
			log.Printf("error: chrome operation has failed: %w", err)
		}
	}()

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

	repo, _ := git.NewDir(dataDir)
	keyring, _ := gpg.NewKeyring("")
	pstore, _ := store.New(repo, keyring)

	h := ChromeHandler{
		dir:     dataDir,
		repo:    repo,
		keyring: keyring,
		pstore:  pstore,
	}
	return h.ServeChrome(context.Background(), os.Stdin, os.Stdout)
}

type ChromeRequest struct {
	CheckStatus *CheckStatusRequest `json:"check_status"`
	CreateKey   *CreateKeyRequest   `json:"create_key"`
	ImportKey   *ImportKeyRequest   `json:"import_key"`
	CreateRepo  *CreateRepoRequest  `json:"create_repo"`
	ImportRepo  *ImportRepoRequest  `json:"import_repo"`
	AddRemote   *AddRemoteRequest   `json:"add_remote"`
	SyncRemote  *SyncRemoteRequest  `json:"sync_remote"`

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

	CheckStatus *CheckStatusResponse `json:"check_status"`
	CreateKey   *CreateKeyResponse   `json:"create_key"`
	ImportKey   *ImportKeyResponse   `json:"import_key"`
	CreateRepo  *CreateRepoResponse  `json:"create_repo"`
	ImportRepo  *ImportRepoResponse  `json:"import_repo"`
	AddRemote   *AddRemoteResponse   `json:"add_remote"`
	SyncRemote  *SyncRemoteResponse  `json:"sync_remote"`

	AddFile    *AddFileResponse    `json:"add_file"`
	EditFile   *EditFileResponse   `json:"edit_file"`
	ListFiles  *ListFilesResponse  `json:"list_files"`
	ViewFile   *ViewFileResponse   `json:"view_file"`
	DeleteFile *DeleteFileResponse `json:"delete_file"`
}

type CheckStatusRequest struct {
}

type CheckStatusResponse struct {
	GPGPath string `json:"gpg_path"`
	GitPath string `json:"git_path"`

	LocalKeys     []*gpg.PublicKey `json:"local_keys"`
	RemoteKeys    []*gpg.PublicKey `json:"remote_keys"`
	ExpiredKeys   []*gpg.PublicKey `json:"expired_keys"`
	UntrustedKeys []*gpg.PublicKey `json:"untrusted_keys"`

	PasswordStoreKeys []*gpg.PublicKey `json:"password_store_keys"`

	Remote string `json:"remote"`
}

type CreateRepoRequest struct {
	Fingerprints []string `json:"fingerprints"`
}

type CreateRepoResponse struct {
}

type ImportRepoRequest struct {
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	Password string `json:"password"`
	Path     string `json:"path"`
}

type ImportRepoResponse struct {
}

type AddRemoteRequest struct {
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	Password string `json:"password"`
	Path     string `json:"path"`
}

type AddRemoteResponse struct {
	SyncRemote *SyncRemoteResponse `json:"sync_remote"`
}

type SyncRemoteRequest struct {
	Fetch bool `json:"fetch"`
	Pull  bool `json:"pull"`
	Push  bool `json:"push"`
}

type SyncRemoteResponse struct {
	Head   *git.LogItem `json:"head"`
	Remote *git.LogItem `json:"remote"`

	NewerCommit string `json:"newer_commit"`
}

type CreateKeyRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Passphrase string `json:"passphrase"`
	KeyLength  int    `json:"key_length,string"`
	KeyYears   int    `json:"key_years,string"`
}

type CreateKeyResponse struct {
}

type ImportKeyRequest struct {
	Key string `json:"key"`
}

type ImportKeyResponse struct {
	NewPublicKeys []*gpg.PublicKey
	NewSecretKeys []*gpg.SecretKey
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
	dir     string
	repo    *git.Dir
	keyring *gpg.Keyring
	pstore  *store.PasswordStore
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

	var resp ChromeResponse
	switch {
	case req.CheckStatus != nil:
		resp.CheckStatus = new(CheckStatusResponse)
		if err := c.doCheckStatus(ctx, req.CheckStatus, resp.CheckStatus); err != nil {
			resp.Status = err.Error()
		}
	case req.CreateKey != nil:
		resp.CreateKey = new(CreateKeyResponse)
		if err := c.doCreateKey(ctx, req.CreateKey, resp.CreateKey); err != nil {
			resp.Status = err.Error()
		}
	case req.ImportKey != nil:
		resp.ImportKey = new(ImportKeyResponse)
		if err := c.doImportKey(ctx, req.ImportKey, resp.ImportKey); err != nil {
			resp.Status = err.Error()
		}
	case req.CreateRepo != nil:
		resp.CreateRepo = new(CreateRepoResponse)
		if err := c.doCreateRepo(ctx, req.CreateRepo, resp.CreateRepo); err != nil {
			resp.Status = err.Error()
		}
	case req.ImportRepo != nil:
		resp.ImportRepo = new(ImportRepoResponse)
		if err := c.doImportRepo(ctx, req.ImportRepo, resp.ImportRepo); err != nil {
			resp.Status = err.Error()
		}
	case req.AddRemote != nil:
		resp.AddRemote = new(AddRemoteResponse)
		if err := c.doAddRemote(ctx, req.AddRemote, resp.AddRemote); err != nil {
			resp.Status = err.Error()
		}
	case req.SyncRemote != nil:
		resp.SyncRemote = new(SyncRemoteResponse)
		if err := c.doSyncRemote(ctx, req.SyncRemote, resp.SyncRemote); err != nil {
			resp.Status = err.Error()
		}
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

func (c *ChromeHandler) doCheckStatus(ctx context.Context, req *CheckStatusRequest, resp *CheckStatusResponse) error {
	if p, err := exec.LookPath("git"); err == nil {
		resp.GitPath = p
	}
	if p, err := exec.LookPath("gpg"); err == nil {
		resp.GPGPath = p
	}
	if c.pstore != nil {
		resp.PasswordStoreKeys = c.pstore.DefaultKeys()
	}
	if c.repo != nil {
		if addr, err := c.repo.GetRemoteURL("past-remote"); err == nil {
			resp.Remote = addr
		}
	}
	if c.keyring != nil {
		now := time.Now()
		// Identify public keys with the private key and others.
		skeys := c.keyring.SecretKeys()
		skeyMap := make(map[string]*gpg.SecretKey)
		for _, skey := range skeys {
			skeyMap[skey.Fingerprint] = skey
		}
		pkeys := c.keyring.PublicKeys()
		for _, pkey := range pkeys {
			if !pkey.CanEncrypt {
				continue
			}
			if !pkey.Trusted {
				resp.UntrustedKeys = append(resp.UntrustedKeys, pkey)
				continue
			}
			if !pkey.ExpiresAt.IsZero() && now.After(pkey.ExpiresAt) {
				resp.ExpiredKeys = append(resp.ExpiredKeys, pkey)
				continue
			}
			if _, ok := skeyMap[pkey.Fingerprint]; ok {
				resp.LocalKeys = append(resp.LocalKeys, pkey)
			} else {
				resp.RemoteKeys = append(resp.RemoteKeys, pkey)
			}
		}
	}
	return nil
}

func (c *ChromeHandler) doCreateKey(ctx context.Context, req *CreateKeyRequest, resp *CreateKeyResponse) error {
	if c.keyring != nil {
		return xerrors.Errorf("gpg keyring already exists: %w", os.ErrInvalid)
	}
	ring, err := gpg.Create(req.Name, req.Email, req.Passphrase, req.KeyLength, req.KeyYears)
	if err != nil {
		return err
	}
	c.keyring = ring
	return nil
}

func (c *ChromeHandler) doImportKey(ctx context.Context, req *ImportKeyRequest, resp *ImportKeyResponse) error {
	if c.keyring == nil {
		return xerrors.Errorf("keyring is not initialized: %w", os.ErrInvalid)
	}
	pkeys, skeys, err := c.keyring.Import([]byte(req.Key))
	if err != nil {
		return xerrors.Errorf("could not import new key(s): %w", err)
	}
	log.Printf("imported %d new public keys and %d new secret keys", len(pkeys), len(skeys))
	resp.NewPublicKeys = pkeys
	resp.NewSecretKeys = skeys
	return nil
}

func (c *ChromeHandler) doCreateRepo(ctx context.Context, req *CreateRepoRequest, resp *CreateRepoResponse) error {
	if c.keyring == nil {
		return xerrors.Errorf("keyring is not initialized: %w", os.ErrInvalid)
	}
	if c.pstore != nil {
		return xerrors.Errorf("git repository already exists: %w", os.ErrInvalid)
	}
	if c.repo == nil {
		repo, err := git.Init(c.dir)
		if err != nil {
			return err
		}
		c.repo = repo
	}
	pstore, err := store.Create(c.repo, c.keyring, req.Fingerprints)
	if err != nil {
		return err
	}
	c.pstore = pstore
	return nil
}

func (c *ChromeHandler) doImportRepo(ctx context.Context, req *ImportRepoRequest, resp *ImportRepoResponse) (status error) {
	if c.keyring == nil {
		return xerrors.Errorf("keyring is not initialized: %w", os.ErrInvalid)
	}

	// TODO: When ssh is chosen, we could verify if ssh passphrase is cached and
	// report an user friendly error.

	reqUsername := url.QueryEscape(req.Username)
	reqPath := filepath.Clean(filepath.Join("/", req.Path))

	remoteURL := ""
	remoteName := "past-remote"
	switch req.Protocol {
	case "ssh":
		remoteURL = fmt.Sprintf("ssh://%s@%s%s", reqUsername, req.Hostname, reqPath)
	case "https":
		remoteURL = fmt.Sprintf("https://%s@%s%s", reqUsername, req.Hostname, reqPath)
	case "git":
		remoteURL = fmt.Sprintf("git://%s@%s%s", reqUsername, req.Hostname, reqPath)
	default:
		return xerrors.Errorf("unsupported git url protocol %q: %w", req.Protocol, os.ErrInvalid)
	}

	repo, err := git.Init(c.dir)
	if err != nil {
		return xerrors.Errorf("could not git init: %w", err)
	}
	defer func() {
		if status != nil {
			if err := os.RemoveAll(c.dir); err != nil {
				log.Panicf("could not remove temporary git directory %q: %w", c.dir, err)
			}
		}
	}()

	// Create credential store file.
	if len(req.Password) > 0 {
		reqPassword := url.QueryEscape(req.Password)
		creds := fmt.Sprintf("%s://%s:%s@%s\n", req.Protocol, reqUsername, reqPassword, req.Hostname)
		credStore := filepath.Join(c.dir, ".past-remote-credentials")
		file, err := os.OpenFile(credStore, os.O_CREATE|os.O_WRONLY, os.FileMode(0600))
		if err != nil {
			return xerrors.Errorf("could not create credential store file %q: %w", credStore, err)
		}
		defer file.Close()
		if _, err := file.Write([]byte(creds)); err != nil {
			return xerrors.Errorf("could not write to credentials file: %w", err)
		}
		// Configure git credential store.
		configValue := fmt.Sprintf("store --file=%s", credStore)
		if err := repo.SetConfg("credential.helper", configValue); err != nil {
			return xerrors.Errorf("could not configure credential store: %w", err)
		}
	}

	if err := repo.AddRemote(remoteName, remoteURL); err != nil {
		return xerrors.Errorf("could not add remote: %w", err)
	}

	if err := repo.FetchAll(); err != nil {
		return xerrors.Errorf("could not git fetch from remotes: %w", err)
	}

	if err := repo.Reset(fmt.Sprintf("%s/master", remoteName)); err != nil {
		return xerrors.Errorf("could not reset working copy to remote master: %w", err)
	}
	return nil
}

func (c *ChromeHandler) doAddRemote(ctx context.Context, req *AddRemoteRequest, resp *AddRemoteResponse) (status error) {
	if c.repo == nil {
		return xerrors.Errorf("git repository is not initialized: %w", os.ErrInvalid)
	}

	reqUsername := url.QueryEscape(req.Username)
	reqPath := filepath.Clean(filepath.Join("/", req.Path))

	remoteURL := ""
	remoteName := "past-remote"
	switch req.Protocol {
	case "ssh":
		remoteURL = fmt.Sprintf("ssh://%s@%s%s", reqUsername, req.Hostname, reqPath)
	case "https":
		remoteURL = fmt.Sprintf("https://%s@%s%s", reqUsername, req.Hostname, reqPath)
	case "git":
		remoteURL = fmt.Sprintf("git://%s@%s%s", reqUsername, req.Hostname, reqPath)
	default:
		return xerrors.Errorf("unsupported git url protocol %q: %w", req.Protocol, os.ErrInvalid)
	}

	if err := c.repo.AddRemote(remoteName, remoteURL); err != nil {
		return xerrors.Errorf("could not add remote %q: %w", remoteName, err)
	}
	defer func() {
		if status != nil {
			if err := c.repo.RemoveRemote(remoteName); err != nil {
				log.Panicf("could not undo adding remote %q: %w", remoteName, err)
			}
		}
	}()

	// Create credential store file. FIXME: Existing credentials file if any will
	// be overwritten.

	if len(req.Password) > 0 {
		reqPassword := url.QueryEscape(req.Password)
		creds := fmt.Sprintf("%s://%s:%s@%s\n", req.Protocol, reqUsername, reqPassword, req.Hostname)
		credStore := filepath.Join(c.dir, ".past-remote-credentials")
		file, err := os.OpenFile(credStore, os.O_CREATE|os.O_WRONLY, os.FileMode(0600))
		if err != nil {
			return xerrors.Errorf("could not create credential store file %q: %w", credStore, err)
		}
		defer file.Close()
		if _, err := file.Write([]byte(creds)); err != nil {
			return xerrors.Errorf("could not write to credentials file: %w", err)
		}
		// Configure git credential store.
		configValue := fmt.Sprintf("store --file=%s", credStore)
		if err := c.repo.SetConfg("credential.helper", configValue); err != nil {
			return xerrors.Errorf("could not configure credential store: %w", err)
		}
		defer func() {
			if status != nil {
				if err := os.Remove(credStore); err != nil {
					log.Printf("error: could not remove credential store: %w", err)
				}
				if err := c.repo.UnsetConfig("credential.helper"); err != nil {
					log.Printf("error: could not unset credential helper: %w", err)
				}
			}
		}()
	}

	syncReq := &SyncRemoteRequest{Fetch: true}
	syncResp := new(SyncRemoteResponse)
	if err := c.doSyncRemote(ctx, syncReq, syncResp); err != nil {
		return xerrors.Errorf("could not determine the diff with remote %q: %w", remoteName, err)
	}
	resp.SyncRemote = syncResp
	data, _ := json.MarshalIndent(resp, "", "  ")
	log.Printf("resp: %s\n", data)
	return nil
}

func (c *ChromeHandler) doSyncRemote(ctx context.Context, req *SyncRemoteRequest, resp *SyncRemoteResponse) error {
	if c.repo == nil {
		return xerrors.Errorf("git repository is not initialized: %w", os.ErrInvalid)
	}
	remoteName := "past-remote"
	remoteMaster := "past-remote/master"
	switch {
	case req.Fetch:
		if err := c.repo.Fetch(remoteName); err != nil {
			return xerrors.Errorf("could not fetch from remote: %w", err)
		}
	case req.Push:
		if err := c.repo.PushOverwrite(remoteName, "master"); err != nil {
			return xerrors.Errorf("could not push to %q: %w", remoteMaster, err)
		}
	case req.Pull:
		if err := c.repo.Reset(remoteMaster); err != nil {
			return xerrors.Errorf("could not pull from %q: %w", remoteMaster, err)
		}
	}
	head, err := c.repo.GetLogItem("HEAD")
	if err != nil {
		return xerrors.Errorf("could not get head log tip: %w", err)
	}
	remote, err := c.repo.GetLogItem(remoteMaster)
	if err != nil {
		return xerrors.Errorf("could not get %q log tip: %w", remoteMaster, err)
	}
	resp.Head = head
	resp.Remote = remote

	if head.Commit != remote.Commit {
		if yes, _ := c.repo.IsAncestor(head.Commit, remote.Commit); yes {
			// Remote master has more commits than head.
			resp.NewerCommit = remote.Commit
		} else if yes, _ := c.repo.IsAncestor(remote.Commit, head.Commit); yes {
			// Head has more commits than remote master.
			resp.NewerCommit = head.Commit
		}
	}

	return nil
}

func (c *ChromeHandler) doAddFile(ctx context.Context, req *AddFileRequest, resp *AddFileResponse) error {
	if c.pstore == nil {
		return xerrors.Errorf("password store is unavailable to add file: %w", os.ErrInvalid)
	}
	if strings.Contains(req.File, "/") {
		return xerrors.Errorf("directories are not allowed in the file name: %w", os.ErrInvalid)
	}

	vs := store.NewValues(nil)
	for _, other := range req.Rest {
		vs.Set(other[0], other[1])
	}
	vs.Set("username", req.Username)
	vs.Set("sitename", req.Sitename)

	data := store.Format(req.Password, vs.Bytes())
	if err := c.pstore.CreateFile(req.File, data, os.FileMode(0644)); err != nil {
		return xerrors.Errorf("could not add new file: %w", err)
	}
	return nil
}

func (c *ChromeHandler) doEditFile(ctx context.Context, req *EditFileRequest, resp *EditFileResponse) error {
	if c.pstore == nil {
		return xerrors.Errorf("password store is unavailable to edit file: %w", os.ErrInvalid)
	}
	if strings.Contains(req.File, "/") {
		return xerrors.Errorf("directories are not allowed in the file name: %w", os.ErrInvalid)
	}

	vs := store.NewValues([]byte(req.Data))
	vs.Set("username", req.Username)
	vs.Set("sitename", req.Sitename)

	data := store.Format(req.Password, vs.Bytes())
	if len(req.OrigFile) > 0 && req.OrigFile != req.File {
		if err := c.pstore.ReplaceFile(req.OrigFile, req.File, data); err != nil {
			return xerrors.Errorf("could not replace file %q: %w", req.OrigFile, err)
		}
	} else {
		if err := c.pstore.UpdateFile(req.File, data); err != nil {
			return xerrors.Errorf("could not update file %q: %w", req.File, err)
		}
	}
	return nil
}

func (c *ChromeHandler) doListFiles(ctx context.Context, req *ListFilesRequest, resp *ListFilesResponse) error {
	if c.pstore == nil {
		return xerrors.Errorf("password store is unavailable to list files (%+v): %w", *c, os.ErrInvalid)
	}

	files, err := c.repo.ListFiles()
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
	if c.pstore == nil {
		return xerrors.Errorf("password store is unavailable to view file: %w", os.ErrInvalid)
	}

	file := filepath.Join("./", req.File+".gpg")
	encrypted, err := c.repo.ReadFile(file)
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
	if c.pstore == nil {
		return xerrors.Errorf("password store is unavailable to delete file: %w", os.ErrInvalid)
	}

	file := filepath.Clean(filepath.Join("./", req.File+".gpg"))
	if err := c.repo.RemoveFile(file); err != nil {
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
