// Copyright (c) 2020 BVK Chaitanya

package msg

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bvk/past"
	"github.com/bvk/past/git"
	"github.com/bvk/past/gpg"

	"golang.org/x/xerrors"
)

func New(dataDir string) (_ *Handler, status error) {
	repo, _ := git.NewDir(dataDir)
	ring, _ := gpg.NewKeyring("")
	pstore, _ := past.New(repo, ring)

	h := &Handler{
		dir:    dataDir,
		repo:   repo,
		ring:   ring,
		pstore: pstore,
	}
	return h, nil
}

type Request struct {
	CheckStatus *CheckStatusRequest `json:"check_status"`

	CreateKey *CreateKeyRequest `json:"create_key"`
	ImportKey *ImportKeyRequest `json:"import_key"`
	EditKey   *EditKeyRequest   `json:"edit_key"`
	ExportKey *ExportKeyRequest `json:"export_key"`
	DeleteKey *DeleteKeyRequest `json:"delete_key"`

	CreateRepo *CreateRepoRequest `json:"create_repo"`
	ImportRepo *ImportRepoRequest `json:"import_repo"`

	AddRemote  *AddRemoteRequest  `json:"add_remote"`
	SyncRemote *SyncRemoteRequest `json:"sync_remote"`

	ScanStore       *ScanStoreRequest       `json:"scan_store"`
	AddRecipient    *AddRecipientRequest    `json:"add_recipient"`
	RemoveRecipient *RemoveRecipientRequest `json:"remove_recipient"`

	AddFile    *AddFileRequest    `json:"add_file"`
	EditFile   *EditFileRequest   `json:"edit_file"`
	ListFiles  *ListFilesRequest  `json:"list_files"`
	ViewFile   *ViewFileRequest   `json:"view_file"`
	DeleteFile *DeleteFileRequest `json:"delete_file"`
}

type Response struct {
	// Status contains any error in performing the operation. It should be empty
	// on success.
	Status string `json:"status"`

	CheckStatus *CheckStatusResponse `json:"check_status"`

	CreateKey *CreateKeyResponse `json:"create_key"`
	ImportKey *ImportKeyResponse `json:"import_key"`
	EditKey   *EditKeyResponse   `json:"edit_key"`
	ExportKey *ExportKeyResponse `json:"export_key"`
	DeleteKey *DeleteKeyResponse `json:"delete_key"`

	CreateRepo *CreateRepoResponse `json:"create_repo"`
	ImportRepo *ImportRepoResponse `json:"import_repo"`

	AddRemote  *AddRemoteResponse  `json:"add_remote"`
	SyncRemote *SyncRemoteResponse `json:"sync_remote"`

	ScanStore       *ScanStoreResponse       `json:"scan_store"`
	AddRecipient    *AddRecipientResponse    `json:"add_recipient"`
	RemoveRecipient *RemoveRecipientResponse `json:"remove_recipient"`

	AddFile    *AddFileResponse    `json:"add_file"`
	EditFile   *EditFileResponse   `json:"edit_file"`
	ListFiles  *ListFilesResponse  `json:"list_files"`
	ViewFile   *ViewFileResponse   `json:"view_file"`
	DeleteFile *DeleteFileResponse `json:"delete_file"`
}

type CheckStatusRequest struct {
	ListFiles bool `json:"list_files"`
}

type CheckStatusResponse struct {
	GPGPath string `json:"gpg_path"`
	GitPath string `json:"git_path"`

	LocalKeys   []*past.PublicKeyData `json:"local_keys"`
	RemoteKeys  []*past.PublicKeyData `json:"remote_keys"`
	ExpiredKeys []*past.PublicKeyData `json:"expired_keys"`

	PasswordStoreKeys []string `json:"password_store_keys"`

	Remote string `json:"remote"`

	ListFiles *ListFilesResponse `json:"list_files"`
}

type CreateRepoRequest struct {
	Fingerprints []string `json:"fingerprints"`

	CheckStatus bool `json:"check_status"`
}

type CreateRepoResponse struct {
	CheckStatus *CheckStatusResponse `json:"check_status"`
}

type ImportRepoRequest struct {
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	Password string `json:"password"`
	Path     string `json:"path"`

	CheckStatus bool `json:"check_status"`
}

type ImportRepoResponse struct {
	CheckStatus *CheckStatusResponse `json:"check_status"`
}

type AddRemoteRequest struct {
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	Password string `json:"password"`
	Path     string `json:"path"`

	SyncRemote bool `json:"sync_remote"`
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

	CheckStatus bool `json:"check_status"`
}

type CreateKeyResponse struct {
	CheckStatus *CheckStatusResponse
}

type ImportKeyRequest struct {
	Key string `json:"key"`

	CheckStatus bool `json:"check_status"`
}

type ImportKeyResponse struct {
	NewPublicKeys []*gpg.PublicKey
	NewSecretKeys []*gpg.SecretKey

	CheckStatus *CheckStatusResponse `json:"check_status"`
}

type EditKeyRequest struct {
	Fingerprint string `json:"fingerprint"`
	Trust       bool   `json:"trust"`
}

type EditKeyResponse struct {
	Key *past.PublicKeyData `json:"key"`
}

type ExportKeyRequest struct {
	Fingerprint string `json:"fingerprint"`
}

type ExportKeyResponse struct {
	ArmorKey string `json:"armor_key"`
}

type DeleteKeyRequest struct {
	Fingerprint string `json:"fingerprint"`

	CheckStatus bool `json:"check_status"`
}

type DeleteKeyResponse struct {
	CheckStatus *CheckStatusResponse `json:"check_status"`
}

type ScanStoreRequest struct {
}

type ScanStoreResponse struct {
	NumFiles int `json:"num_files"`

	KeyMap       map[string]*past.PublicKeyData `json:"key_map"`
	UnusedKeyMap map[string]*past.PublicKeyData `json:"unused_key_map"`

	KeyFileCountMap        map[string]int `json:"key_file_count_map"`
	MissingKeyFileCountMap map[string]int `json:"missing_key_file_count_map"`
}

type AddRecipientRequest struct {
	NumSkip     int    `json:"num_skip"`
	Fingerprint string `json:"fingerprint"`
	ScanStore   bool   `json:"scan_store"`
}

type AddRecipientResponse struct {
	ScanStore *ScanStoreResponse `json:"scan_store"`
}

type RemoveRecipientRequest struct {
	NumSkip     int    `json:"num_skip"`
	Fingerprint string `json:"fingerprint"`
	ScanStore   bool   `json:"scan_store"`
}

type RemoveRecipientResponse struct {
	ScanStore *ScanStoreResponse `json:"scan_store"`
}

type ListFilesRequest struct {
}

type ListFilesResponse struct {
	Files []string `json:"files"`
}

type AddFileRequest struct {
	Filename string      `json:"filename"`
	Password string      `json:"password"`
	Username string      `json:"username"`
	Sitename string      `json:"sitename"`
	Data     string      `json:"data"`
	Rest     [][2]string `json:"rest"`
}

type AddFileResponse struct {
}

type EditFileRequest struct {
	OrigFile string `json:"orig_file"`

	Filename string `json:"filename"`
	Password string `json:"password"`
	Username string `json:"username"`
	Sitename string `json:"sitename"`
	Data     string `json:"data"`
}

type EditFileResponse struct {
}

type ViewFileRequest struct {
	Filename string `json:"filename"`
}

type ViewFileResponse struct {
	Filename string `json:"filename"`
	Sitename string `json:"sitename"`
	Username string `json:"username"`
	Password string `json:"password"`
	Data     string `json:"data"`
}

type DeleteFileRequest struct {
	File string `json:"file"`
}

type DeleteFileResponse struct {
}

type Handler struct {
	dir    string
	repo   *git.Dir
	ring   *gpg.Keyring
	pstore *past.PasswordStore
}

func (h *Handler) Serve(ctx context.Context, in io.Reader, out io.Writer) (status error) {
	defer func() {
		if status != nil {
			log.Printf("error: browser operation has failed: %w", status)
		}
	}()

	var sizeBytes [4]byte
	if _, err := io.ReadFull(os.Stdin, sizeBytes[:]); err != nil {
		return xerrors.Errorf("could not read input message length: %w", err)
	}
	size := binary.LittleEndian.Uint32(sizeBytes[:])
	reqBuf := make([]byte, size)
	if _, err := io.ReadFull(os.Stdin, reqBuf); err != nil {
		return xerrors.Errorf("could not read input message: %w", err)
	}

	req := new(Request)
	if err := json.Unmarshal(reqBuf, req); err != nil {
		return xerrors.Errorf("could not unmarshal input message: %w", err)
	}

	resp, err := h.handleRequest(ctx, req)
	if err != nil {
		return xerrors.Errorf("could not handle input request: %w", err)
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

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	req := new(Request)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, fmt.Sprintf("could not unmarshal input message: %v", err), http.StatusInternalServerError)
		return
	}

	resp, err := h.handleRequest(r.Context(), req)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not handle input request: %v", err), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleRequest(ctx context.Context, req *Request) (_ *Response, status error) {
	resp := new(Response)
	defer func() {
		if len(resp.Status) > 0 {
			log.Printf("error: operation has failed with response status: %s", resp.Status)
		}
	}()

	switch {
	case req.CheckStatus != nil:
		resp.CheckStatus = new(CheckStatusResponse)
		if err := h.doCheckStatus(ctx, req.CheckStatus, resp.CheckStatus); err != nil {
			resp.Status = err.Error()
		}
	case req.CreateKey != nil:
		resp.CreateKey = new(CreateKeyResponse)
		if err := h.doCreateKey(ctx, req.CreateKey, resp.CreateKey); err != nil {
			resp.Status = err.Error()
		}
	case req.ImportKey != nil:
		resp.ImportKey = new(ImportKeyResponse)
		if err := h.doImportKey(ctx, req.ImportKey, resp.ImportKey); err != nil {
			resp.Status = err.Error()
		}
	case req.EditKey != nil:
		resp.EditKey = new(EditKeyResponse)
		if err := h.doEditKey(ctx, req.EditKey, resp.EditKey); err != nil {
			resp.Status = err.Error()
		}
	case req.ExportKey != nil:
		resp.ExportKey = new(ExportKeyResponse)
		if err := h.doExportKey(ctx, req.ExportKey, resp.ExportKey); err != nil {
			resp.Status = err.Error()
		}
	case req.DeleteKey != nil:
		resp.DeleteKey = new(DeleteKeyResponse)
		if err := h.doDeleteKey(ctx, req.DeleteKey, resp.DeleteKey); err != nil {
			resp.Status = err.Error()
		}
	case req.CreateRepo != nil:
		resp.CreateRepo = new(CreateRepoResponse)
		if err := h.doCreateRepo(ctx, req.CreateRepo, resp.CreateRepo); err != nil {
			resp.Status = err.Error()
		}
	case req.ImportRepo != nil:
		resp.ImportRepo = new(ImportRepoResponse)
		if err := h.doImportRepo(ctx, req.ImportRepo, resp.ImportRepo); err != nil {
			resp.Status = err.Error()
		}
	case req.AddRemote != nil:
		resp.AddRemote = new(AddRemoteResponse)
		if err := h.doAddRemote(ctx, req.AddRemote, resp.AddRemote); err != nil {
			resp.Status = err.Error()
		}
	case req.SyncRemote != nil:
		resp.SyncRemote = new(SyncRemoteResponse)
		if err := h.doSyncRemote(ctx, req.SyncRemote, resp.SyncRemote); err != nil {
			resp.Status = err.Error()
		}
	case req.ScanStore != nil:
		resp.ScanStore = new(ScanStoreResponse)
		if err := h.doScanStore(ctx, req.ScanStore, resp.ScanStore); err != nil {
			resp.Status = err.Error()
		}
	case req.AddRecipient != nil:
		resp.AddRecipient = new(AddRecipientResponse)
		if err := h.doAddRecipient(ctx, req.AddRecipient, resp.AddRecipient); err != nil {
			resp.Status = err.Error()
		}
	case req.RemoveRecipient != nil:
		resp.RemoveRecipient = new(RemoveRecipientResponse)
		if err := h.doRemoveRecipient(ctx, req.RemoveRecipient, resp.RemoveRecipient); err != nil {
			resp.Status = err.Error()
		}
	case req.AddFile != nil:
		resp.AddFile = new(AddFileResponse)
		if err := h.doAddFile(ctx, req.AddFile, resp.AddFile); err != nil {
			resp.Status = err.Error()
		}
	case req.EditFile != nil:
		resp.EditFile = new(EditFileResponse)
		if err := h.doEditFile(ctx, req.EditFile, resp.EditFile); err != nil {
			resp.Status = err.Error()
		}
	case req.ListFiles != nil:
		resp.ListFiles = new(ListFilesResponse)
		if err := h.doListFiles(ctx, req.ListFiles, resp.ListFiles); err != nil {
			resp.Status = err.Error()
		}
	case req.ViewFile != nil:
		resp.ViewFile = new(ViewFileResponse)
		if err := h.doViewFile(ctx, req.ViewFile, resp.ViewFile); err != nil {
			resp.Status = err.Error()
		}
	case req.DeleteFile != nil:
		resp.DeleteFile = new(DeleteFileResponse)
		if err := h.doDeleteFile(ctx, req.DeleteFile, resp.DeleteFile); err != nil {
			resp.Status = err.Error()
		}
	default:
		resp.Status = xerrors.Errorf("unknown or invalid request: %w", os.ErrInvalid).Error()
	}

	return resp, nil
}

func getPublicKeysData(ring *gpg.Keyring) []*past.PublicKeyData {
	// Identify public keys with the private key and others.
	skeyMap := make(map[string]*gpg.SecretKey)
	for _, skey := range ring.SecretKeys() {
		skeyMap[skey.Fingerprint] = skey
	}
	var pks []*past.PublicKeyData
	for _, pkey := range ring.PublicKeys() {
		if !pkey.CanEncrypt {
			continue
		}
		if skey, ok := skeyMap[pkey.Fingerprint]; ok {
			pks = append(pks, past.ToPublicKeyData(pkey, skey))
		} else {
			pks = append(pks, past.ToPublicKeyData(pkey, nil))
		}
	}
	return pks
}

func (h *Handler) doCheckStatus(ctx context.Context, req *CheckStatusRequest, resp *CheckStatusResponse) error {
	if p, err := exec.LookPath("git"); err == nil {
		resp.GitPath = p
	}
	if p, err := exec.LookPath("gpg"); err == nil {
		resp.GPGPath = p
	}
	if repo, err := git.NewDir(h.dir); err == nil {
		h.repo = repo
	}
	if h.repo != nil {
		if addr, err := h.repo.GetRemoteURL("past-remote"); err == nil {
			resp.Remote = addr
		}
	}
	if ring, err := gpg.NewKeyring(""); err == nil {
		h.ring = ring
	}
	if h.ring != nil {
		// Identify public keys with the private key and others.
		skeys := h.ring.SecretKeys()
		skeyMap := make(map[string]*gpg.SecretKey)
		for _, skey := range skeys {
			skeyMap[skey.Fingerprint] = skey
		}
		pkeys := h.ring.PublicKeys()
		for _, pkey := range pkeys {
			if !pkey.CanEncrypt {
				continue
			}
			var v *past.PublicKeyData
			if skey, ok := skeyMap[pkey.Fingerprint]; ok {
				v = past.ToPublicKeyData(pkey, skey)
			} else {
				v = past.ToPublicKeyData(pkey, nil)
			}
			if v.CanDecrypt {
				resp.LocalKeys = append(resp.LocalKeys, v)
			} else if v.IsExpired {
				resp.ExpiredKeys = append(resp.ExpiredKeys, v)
			} else {
				resp.RemoteKeys = append(resp.RemoteKeys, v)
			}
		}
	}
	if pstore, err := past.New(h.repo, h.ring); err == nil {
		h.pstore = pstore
	}
	if h.pstore != nil {
		resp.PasswordStoreKeys, _ = h.pstore.FileKeys(".")
	}
	if h.pstore != nil && req.ListFiles {
		lsReq := new(ListFilesRequest)
		lsResp := new(ListFilesResponse)
		if err := h.doListFiles(ctx, lsReq, lsResp); err != nil {
			return xerrors.Errorf("could not list files after the status check: %w", err)
		}
		resp.ListFiles = lsResp
	}
	return nil
}

func (h *Handler) doCreateKey(ctx context.Context, req *CreateKeyRequest, resp *CreateKeyResponse) error {
	if _, err := gpg.Create(req.Name, req.Email, req.Passphrase, req.KeyLength, req.KeyYears); err != nil {
		return err
	}
	if req.CheckStatus {
		csReq := new(CheckStatusRequest)
		csResp := new(CheckStatusResponse)
		if err := h.doCheckStatus(ctx, csReq, csResp); err != nil {
			return xerrors.Errorf("could not check status after creating the key: %w", err)
		}
		resp.CheckStatus = csResp
	}
	return nil
}

func (h *Handler) doImportKey(ctx context.Context, req *ImportKeyRequest, resp *ImportKeyResponse) error {
	if h.ring == nil {
		return xerrors.Errorf("keyring is not initialized: %w", os.ErrInvalid)
	}
	pkeys, skeys, err := h.ring.Import([]byte(req.Key))
	if err != nil {
		return xerrors.Errorf("could not import new key(s): %w", err)
	}
	log.Printf("imported %d new public keys and %d new secret keys", len(pkeys), len(skeys))
	resp.NewPublicKeys = pkeys
	resp.NewSecretKeys = skeys
	if req.CheckStatus {
		csReq := new(CheckStatusRequest)
		csResp := new(CheckStatusResponse)
		if err := h.doCheckStatus(ctx, csReq, csResp); err != nil {
			return xerrors.Errorf("could not check status after creating the key: %w", err)
		}
		resp.CheckStatus = csResp
	}
	return nil
}

func (h *Handler) doEditKey(ctx context.Context, req *EditKeyRequest, resp *EditKeyResponse) error {
	if h.ring == nil {
		return xerrors.Errorf("keyring is not initialized: %w", os.ErrInvalid)
	}
	if err := h.ring.SetTrusted(req.Fingerprint, req.Trust); err != nil {
		return xerrors.Errorf("could not update key trust status: %w", err)
	}
	if err := h.ring.Refresh(); err != nil {
		return xerrors.Errorf("could not refresh keying: %w", err)
	}
	for _, pkey := range getPublicKeysData(h.ring) {
		if pkey.KeyFingerprint == req.Fingerprint {
			resp.Key = pkey
			return nil
		}
	}
	return xerrors.Errorf("could not find the key %q after updating the trust status: %w", req.Fingerprint, os.ErrInvalid)
}

func (h *Handler) doExportKey(ctx context.Context, req *ExportKeyRequest, resp *ExportKeyResponse) error {
	if h.ring == nil {
		return xerrors.Errorf("keyring is not initialized: %w", os.ErrInvalid)
	}
	data, err := h.ring.Export(req.Fingerprint)
	if err != nil {
		return xerrors.Errorf("could not export key %q: %w", req.Fingerprint, err)
	}
	resp.ArmorKey = string(data)
	return nil
}

func (h *Handler) doDeleteKey(ctx context.Context, req *DeleteKeyRequest, resp *DeleteKeyResponse) error {
	if h.ring == nil {
		return xerrors.Errorf("keyring is not initialized: %w", os.ErrInvalid)
	}
	skeys := h.ring.SecretKeys()
	for _, skey := range skeys {
		if skey.Fingerprint == req.Fingerprint {
			if err := h.ring.DeleteSecretKey(req.Fingerprint); err != nil {
				return xerrors.Errorf("could not delete secret key %q: %w", req.Fingerprint, err)
			}
			log.Printf("secret key for %q is deleted successfully", req.Fingerprint)
			break
		}
	}
	if err := h.ring.Delete(req.Fingerprint); err != nil {
		return xerrors.Errorf("could not delete key %q: %w", req.Fingerprint, err)
	}
	if req.CheckStatus {
		csReq := new(CheckStatusRequest)
		csResp := new(CheckStatusResponse)
		if err := h.doCheckStatus(ctx, csReq, csResp); err != nil {
			return xerrors.Errorf("could not check status after deleting the key: %w", err)
		}
		resp.CheckStatus = csResp
	}
	return nil
}

func (h *Handler) doCreateRepo(ctx context.Context, req *CreateRepoRequest, resp *CreateRepoResponse) error {
	if h.ring == nil {
		return xerrors.Errorf("keyring is not initialized: %w", os.ErrInvalid)
	}
	if h.pstore != nil {
		return xerrors.Errorf("git repository already exists: %w", os.ErrInvalid)
	}
	if h.repo == nil {
		repo, err := git.Init(h.dir)
		if err != nil {
			return err
		}
		h.repo = repo
	}
	pstore, err := past.Create(h.repo, h.ring, req.Fingerprints)
	if err != nil {
		return err
	}
	h.pstore = pstore
	return nil
}

func (h *Handler) doImportRepo(ctx context.Context, req *ImportRepoRequest, resp *ImportRepoResponse) (status error) {
	if h.ring == nil {
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

	repo, err := git.Init(h.dir)
	if err != nil {
		return xerrors.Errorf("could not git init: %w", err)
	}
	defer func() {
		if status != nil {
			if err := os.RemoveAll(h.dir); err != nil {
				log.Panicf("could not remove temporary git directory %q: %w", h.dir, err)
			}
		}
	}()

	// Create credential store file.
	if len(req.Password) > 0 {
		reqPassword := url.QueryEscape(req.Password)
		creds := fmt.Sprintf("%s://%s:%s@%s\n", req.Protocol, reqUsername, reqPassword, req.Hostname)
		credStore := filepath.Join(h.dir, ".past-remote-credentials")
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

	// TODO: Check that at least one file can be decrypted with the local keyring.

	return nil
}

func (h *Handler) doAddRemote(ctx context.Context, req *AddRemoteRequest, resp *AddRemoteResponse) (status error) {
	if h.repo == nil {
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

	if err := h.repo.AddRemote(remoteName, remoteURL); err != nil {
		return xerrors.Errorf("could not add remote %q: %w", remoteName, err)
	}
	defer func() {
		if status != nil {
			if err := h.repo.RemoveRemote(remoteName); err != nil {
				log.Panicf("could not undo adding remote %q: %w", remoteName, err)
			}
		}
	}()

	// Create credential store file. FIXME: Existing credentials file if any will
	// be overwritten.

	if len(req.Password) > 0 {
		reqPassword := url.QueryEscape(req.Password)
		creds := fmt.Sprintf("%s://%s:%s@%s\n", req.Protocol, reqUsername, reqPassword, req.Hostname)
		credStore := filepath.Join(h.dir, ".past-remote-credentials")
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
		if err := h.repo.SetConfg("credential.helper", configValue); err != nil {
			return xerrors.Errorf("could not configure credential store: %w", err)
		}
		defer func() {
			if status != nil {
				if err := os.Remove(credStore); err != nil {
					log.Printf("error: could not remove credential store: %w", err)
				}
				if err := h.repo.UnsetConfig("credential.helper"); err != nil {
					log.Printf("error: could not unset credential helper: %w", err)
				}
			}
		}()
	}

	syncReq := &SyncRemoteRequest{Fetch: true}
	syncResp := new(SyncRemoteResponse)
	if err := h.doSyncRemote(ctx, syncReq, syncResp); err != nil {
		return xerrors.Errorf("could not determine the diff with remote %q: %w", remoteName, err)
	}
	resp.SyncRemote = syncResp
	return nil
}

func (h *Handler) doSyncRemote(ctx context.Context, req *SyncRemoteRequest, resp *SyncRemoteResponse) error {
	if h.repo == nil {
		return xerrors.Errorf("git repository is not initialized: %w", os.ErrInvalid)
	}
	remoteName := "past-remote"
	remoteMaster := "past-remote/master"
	switch {
	case req.Fetch:
		if err := h.repo.Fetch(remoteName); err != nil {
			return xerrors.Errorf("could not fetch from remote: %w", err)
		}
	case req.Push:
		if err := h.repo.PushOverwrite(remoteName, "master"); err != nil {
			return xerrors.Errorf("could not push to %q: %w", remoteMaster, err)
		}
	case req.Pull:
		if err := h.repo.Reset(remoteMaster); err != nil {
			return xerrors.Errorf("could not pull from %q: %w", remoteMaster, err)
		}
	}
	head, err := h.repo.GetLogItem("HEAD")
	if err != nil {
		return xerrors.Errorf("could not get head log tip: %w", err)
	}
	remote, err := h.repo.GetLogItem(remoteMaster)
	if err != nil {
		return xerrors.Errorf("could not get %q log tip: %w", remoteMaster, err)
	}
	resp.Head = head
	resp.Remote = remote

	if head.Commit != remote.Commit {
		if yes, _ := h.repo.IsAncestor(head.Commit, remote.Commit); yes {
			// Remote master has more commits than head.
			resp.NewerCommit = remote.Commit
		} else if yes, _ := h.repo.IsAncestor(remote.Commit, head.Commit); yes {
			// Head has more commits than remote master.
			resp.NewerCommit = head.Commit
		}
	}

	return nil
}

func (h *Handler) doScanStore(ctx context.Context, req *ScanStoreRequest, resp *ScanStoreResponse) error {
	if h.pstore == nil {
		return xerrors.Errorf("password store is not initialized: %w", os.ErrInvalid)
	}
	if h.ring == nil {
		return xerrors.Errorf("key ring is not initialized: %w", os.ErrInvalid)
	}
	resp.KeyMap = make(map[string]*past.PublicKeyData)
	resp.UnusedKeyMap = make(map[string]*past.PublicKeyData)

	resp.KeyFileCountMap = make(map[string]int)
	resp.MissingKeyFileCountMap = make(map[string]int)
	// Scan the recipient key ids for each encrypted file, so that we can report
	// number of files accessible to each key id (including the missing key ids).
	files, err := h.pstore.ListFiles()
	if err != nil {
		return xerrors.Errorf("could not list files in the password store: %w", err)
	}
	resp.NumFiles = len(files)
	for _, file := range files {
		rs, err := h.pstore.Recipients(file)
		if err != nil {
			return xerrors.Errorf("couldn't determine recipients for file %q: %w", file, err)
		}
		for _, r := range rs {
			pkey, err := h.pstore.RecipientPublicKey(r)
			if err != nil {
				if !xerrors.Is(err, os.ErrNotExist) {
					return xerrors.Errorf("could not determine public key for %q: %w", r, err)
				}
				resp.MissingKeyFileCountMap[r]++
				continue
			}
			resp.KeyMap[pkey.KeyFingerprint] = pkey
			resp.KeyFileCountMap[pkey.KeyFingerprint]++
		}
	}
	for _, pkey := range getPublicKeysData(h.ring) {
		if _, ok := resp.KeyMap[pkey.KeyFingerprint]; !ok {
			resp.UnusedKeyMap[pkey.KeyFingerprint] = pkey
		}
	}
	return nil
}

func (h *Handler) doAddRecipient(ctx context.Context, req *AddRecipientRequest, resp *AddRecipientResponse) error {
	if h.pstore == nil {
		return xerrors.Errorf("password store is not initialized: %w", os.ErrInvalid)
	}
	keys, err := h.pstore.FileKeys(".")
	if err != nil {
		return xerrors.Errorf("could not determine current gpg ids: %w", err)
	}
	for _, key := range keys {
		if key == req.Fingerprint {
			return xerrors.Errorf("key %q is already a recipient: %w", key, os.ErrExist)
		}
	}
	keys = append(keys, req.Fingerprint)
	if err := h.pstore.Reinit("", keys, req.NumSkip); err != nil {
		return xerrors.Errorf("could not reinitialize with a recipient add: %w", err)
	}
	// Repo and pstore must be updated.
	if repo, err := git.NewDir(h.dir); err != nil {
		return xerrors.Errorf("could not refresh git repo after adding recipient: %w", err)
	} else {
		h.repo = repo
	}
	if pstore, err := past.New(h.repo, h.ring); err != nil {
		return xerrors.Errorf("could not refresh past instance after adding recipient: %w", err)
	} else {
		h.pstore = pstore
	}
	ssReq := new(ScanStoreRequest)
	ssResp := new(ScanStoreResponse)
	if err := h.doScanStore(ctx, ssReq, ssResp); err != nil {
		return xerrors.Errorf("could not scan store after adding recipient: %w", err)
	}
	resp.ScanStore = ssResp
	return nil
}

func (h *Handler) doRemoveRecipient(ctx context.Context, req *RemoveRecipientRequest, resp *RemoveRecipientResponse) error {
	if h.pstore == nil {
		return xerrors.Errorf("password store is not initialized: %w", os.ErrInvalid)
	}
	keys, err := h.pstore.FileKeys(".")
	if err != nil {
		return xerrors.Errorf("could not determine current gpg ids: %w", err)
	}
	var newKeys []string
	if len(req.Fingerprint) > 0 {
		for _, key := range keys {
			if key == req.Fingerprint {
				continue
			}
			newKeys = append(newKeys, key)
		}
	}
	if len(newKeys) == len(keys) {
		return xerrors.Errorf("key %q is not a recipient: %w", req.Fingerprint, os.ErrExist)
	}
	// TODO: We should add support for directories.
	if len(newKeys) > 0 {
		if err := h.pstore.Reinit("", newKeys, req.NumSkip); err != nil {
			return xerrors.Errorf("could not reinitialize with a recipient removed: %w", err)
		}
	} else {
		if err := h.pstore.Reinit("", nil, req.NumSkip); err != nil {
			return xerrors.Errorf("could not reinitialize with a recipient removed: %w", err)
		}
	}
	// Repo and pstore must be updated.
	if repo, err := git.NewDir(h.dir); err != nil {
		return xerrors.Errorf("could not refresh git repo after adding recipient: %w", err)
	} else {
		h.repo = repo
	}
	if pstore, err := past.New(h.repo, h.ring); err != nil {
		return xerrors.Errorf("could not refresh past instance after adding recipient: %w", err)
	} else {
		h.pstore = pstore
	}
	ssReq := new(ScanStoreRequest)
	ssResp := new(ScanStoreResponse)
	if err := h.doScanStore(ctx, ssReq, ssResp); err != nil {
		return xerrors.Errorf("could not scan store after removing recipient: %w", err)
	}
	resp.ScanStore = ssResp
	return nil
}

func (h *Handler) doAddFile(ctx context.Context, req *AddFileRequest, resp *AddFileResponse) error {
	if h.pstore == nil {
		return xerrors.Errorf("password store is unavailable to add file: %w", os.ErrInvalid)
	}

	vs := past.NewValues(nil)
	for _, other := range req.Rest {
		vs.Set(other[0], other[1])
	}
	vs.Set("username", req.Username)
	vs.Set("sitename", req.Sitename)

	if len(req.Filename) == 0 {
		req.Filename = filepath.Join("./", req.Sitename, req.Username)
	}

	data := past.Format(req.Password, vs.Bytes())
	if err := h.pstore.CreateFile(req.Filename, data, os.FileMode(0644)); err != nil {
		return xerrors.Errorf("could not add new file: %w", err)
	}
	return nil
}

func (h *Handler) doEditFile(ctx context.Context, req *EditFileRequest, resp *EditFileResponse) error {
	if h.pstore == nil {
		return xerrors.Errorf("password store is unavailable to edit file: %w", os.ErrInvalid)
	}

	vs := past.NewValues([]byte(req.Data))
	vs.Set("username", req.Username)
	vs.Set("sitename", req.Sitename)

	if len(req.Filename) == 0 {
		req.Filename = filepath.Join("./", req.Sitename, req.Username)
		vs.Del("username")
		vs.Del("sitename")
	}

	data := past.Format(req.Password, vs.Bytes())
	if len(req.OrigFile) > 0 && req.OrigFile != req.Filename {
		if err := h.pstore.ReplaceFile(req.OrigFile, req.Filename, data); err != nil {
			return xerrors.Errorf("could not replace file %q: %w", req.OrigFile, err)
		}
	} else {
		if err := h.pstore.UpdateFile(req.Filename, data); err != nil {
			return xerrors.Errorf("could not update file %q: %w", req.Filename, err)
		}
	}
	return nil
}

func (h *Handler) doListFiles(ctx context.Context, req *ListFilesRequest, resp *ListFilesResponse) error {
	if h.pstore == nil {
		return xerrors.Errorf("password store is not initialized to list files: %w", os.ErrInvalid)
	}

	files, err := h.repo.ListFiles()
	if err != nil {
		return xerrors.Errorf("could not list files in the git directory: %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file, ".gpg") {
			resp.Files = append(resp.Files, strings.TrimSuffix(file, ".gpg"))
		}
	}
	return nil
}

func (h *Handler) doViewFile(ctx context.Context, req *ViewFileRequest, resp *ViewFileResponse) error {
	if h.pstore == nil {
		return xerrors.Errorf("password store is unavailable to view file: %w", os.ErrInvalid)
	}

	file := filepath.Join("./", req.Filename+".gpg")
	encrypted, err := h.repo.ReadFile(file)
	if err != nil {
		return xerrors.Errorf("could not get login entry with name %q: %w", req.Filename, err)
	}
	decrypted, err := h.ring.Decrypt(encrypted)
	if err != nil {
		return xerrors.Errorf("could not decrypt login entry %q: %w", req.Filename, err)
	}
	password, data := past.Parse(decrypted)
	values := past.NewValues(data)

	// Sitename and username are chosen from the filepath by default if the file
	// path is in `site.com/user.gpg` format. However, key-value pairs in the
	// file data can override the default username and sitename values.
	sitename := ""
	dir := filepath.Dir(req.Filename)
	if dir != "." && strings.ContainsRune(dir, '.') && !strings.ContainsRune(dir, filepath.Separator) {
		sitename = dir
	}
	if s := values.Get("sitename"); len(s) > 0 {
		sitename = s
	}

	username := filepath.Base(req.Filename)
	if us := past.GetUsernames(values); len(us) > 0 {
		username = us[0]
	}

	resp.Data = string(data)
	resp.Sitename = sitename
	resp.Username = username
	resp.Password = password
	resp.Filename = req.Filename
	return nil
}

func (h *Handler) doDeleteFile(ctx context.Context, req *DeleteFileRequest, resp *DeleteFileResponse) error {
	if h.pstore == nil {
		return xerrors.Errorf("password store is unavailable to delete file: %w", os.ErrInvalid)
	}

	file := filepath.Clean(filepath.Join("./", req.File+".gpg"))
	if err := h.repo.RemoveFile(file); err != nil {
		return xerrors.Errorf("could not remove file %q: %w", file, err)
	}
	return nil
}
