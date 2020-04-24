// Copyright (c) 2020 BVK Chaitanya

package gpg

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/xerrors"

	"github.com/bvk/past/gpg/internal"
)

type Keyring struct {
	keyring string

	keyRecords  []*internal.Record
	skeyRecords []*internal.Record
}

func NewKeyring(path string) (*Keyring, error) {
	g := &Keyring{keyring: path}
	if err := g.Refresh(); err != nil {
		return nil, xerrors.Errorf("could not list gpg keys: %w", err)
	}
	if keys := g.PublicKeys(); len(keys) == 0 {
		return nil, os.ErrNotExist
	}
	return g, nil
}

func Create(name, email, passphrase string, length, years int) (*Keyring, error) {
	// Kill any existing gpg agent to avoid the following error:
	//
	// gpg: agent_genkey failed: No such file or directory
	//
	exec.Command("gpgconf", "--kill", "gpg-agent").Run()

	file, err := ioutil.TempFile("", "genkeyscript")
	if err != nil {
		return nil, xerrors.Errorf("could not create temp file: %w", err)
	}
	defer func() {
		name := file.Name()
		if err := os.Remove(name); err != nil {
			log.Printf("error: could not remove temporary file %q: %w", name, err)
		}
	}()

	// Key expiry date in years.
	expireDate := "0"
	if years > 0 {
		expireDate = fmt.Sprintf("%dy", years)
	}

	content := `
Key-Type: default
Key-Length: ` + fmt.Sprintf("%d", length) + `
Subkey-Type: default
Subkey-Length: ` + fmt.Sprintf("%d", length) + `
Name-Real: ` + name + `
Name-Email: ` + email + `
Expire-Date: ` + expireDate + `
Passphrase: ` + passphrase + `
`
	if _, err := file.Write([]byte(content)); err != nil {
		return nil, xerrors.Errorf("could not write to temporary file: %w", err)
	}

	cmd := exec.Command("gpg", "--batch", "--gen-key", file.Name())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, xerrors.Errorf("could not create initial key (stderr: %s): %w", stderr.String(), err)
	}
	return NewKeyring("")
}

func (g *Keyring) options() []string {
	opts := []string{
		"--yes",
		"--batch",
		"--keyid-format", "long",
		"--no-greeting",
		"--with-colons",
	}
	if len(g.keyring) > 0 {
		opts = append(opts, "--no-keyring", "--keyring", g.keyring)
	}
	return opts
}

func (g *Keyring) Refresh() error {
	keyCmd := exec.Command("gpg", "--list-keys")
	keyCmd.Args = append(keyCmd.Args, g.options()...)
	keyStdout := &bytes.Buffer{}
	keyCmd.Stdout = keyStdout
	if err := keyCmd.Run(); err != nil {
		return xerrors.Errorf("could not list keys: %w", err)
	}
	keyRecords, err := g.parseRecords(keyStdout.String())
	if err != nil {
		return xerrors.Errorf("could not parse list keys output: %w", err)
	}

	skeyCmd := exec.Command("gpg", "--list-secret-keys")
	skeyCmd.Args = append(skeyCmd.Args, g.options()...)
	skeyStdout := &bytes.Buffer{}
	skeyCmd.Stdout = skeyStdout
	if err := skeyCmd.Run(); err != nil {
		return xerrors.Errorf("could not list secret keys: %w", err)
	}
	skeyRecords, err := g.parseRecords(skeyStdout.String())
	if err != nil {
		return xerrors.Errorf("could not parse list secret keys output: %w", err)
	}
	g.keyRecords = keyRecords
	g.skeyRecords = skeyRecords
	return nil
}

func (g *Keyring) Decrypt(data []byte) ([]byte, error) {
	opts := []string{
		"--compress-algo=none",
		"--no-encrypt-to",
	}
	var buffer bytes.Buffer
	cmd := exec.Command("gpg", "-d")
	cmd.Args = append(cmd.Args, opts...)
	cmd.Args = append(cmd.Args, g.options()...)
	cmd.Stdin = bytes.NewBuffer(data)
	cmd.Stdout = &buffer
	cmd.Stderr = ioutil.Discard
	if err := cmd.Run(); err != nil {
		return nil, xerrors.Errorf("could not decrypt input data: %w", err)
	}
	return buffer.Bytes(), nil
}

func (g *Keyring) Encrypt(data []byte, fps []string) ([]byte, error) {
	opts := []string{
		"--compress-algo=none",
	}
	for _, fp := range fps {
		opts = append(opts, "--recipient", fp)
	}
	cmd := exec.Command("gpg", "--encrypt")
	cmd.Args = append(cmd.Args, opts...)
	cmd.Args = append(cmd.Args, g.options()...)
	cmd.Stdin = bytes.NewBuffer(data)
	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, xerrors.Errorf("could not encrypt input data: %w", err)
	}
	return stdout.Bytes(), nil
}

func (g *Keyring) Delete(fingerprint string) error {
	cmd := exec.Command("gpg", "--delete-keys", "--yes")
	cmd.Args = append(cmd.Args, g.options()...)
	cmd.Args = append(cmd.Args, fingerprint)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Printf("error: delete key cmd %v failed with stderr %q", cmd.Args, stderr.String())
		return xerrors.Errorf("could not delete key %q: %w", fingerprint, err)
	}
	log.Printf("key %q is deleted from the keyring", fingerprint)
	return nil
}

func (g *Keyring) DeleteSecretKey(fingerprint string) error {
	cmd := exec.Command("gpg", "--delete-secret-keys", "--yes")
	cmd.Args = append(cmd.Args, g.options()...)
	cmd.Args = append(cmd.Args, fingerprint)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Printf("error: delete secret key cmd %v failed with stderr %q", cmd.Args, stderr.String())
		return xerrors.Errorf("could not delete secret key %q: %w", fingerprint, err)
	}
	log.Printf("secret key %q is deleted from the keyring", fingerprint)
	return nil
}

func (g *Keyring) Export(fingerprint string) ([]byte, error) {
	cmd := exec.Command("gpg", "--export", "--armor")
	cmd.Args = append(cmd.Args, g.options()...)
	cmd.Args = append(cmd.Args, fingerprint)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		log.Printf("error: export key cmd %v failed with stderr %q", cmd.Args, stderr.String())
		return nil, xerrors.Errorf("could not export key %q: %w", fingerprint, err)
	}
	return stdout.Bytes(), nil
}

func (g *Keyring) Import(key []byte) ([]*PublicKey, []*SecretKey, error) {
	oldPkeys := g.PublicKeys()
	oldPublicKeys := make(map[string]*PublicKey)
	for _, pkey := range oldPkeys {
		oldPublicKeys[pkey.Fingerprint] = pkey
	}
	oldSkeys := g.SecretKeys()
	oldSecretKeys := make(map[string]*SecretKey)
	for _, skey := range oldSkeys {
		oldSecretKeys[skey.Fingerprint] = skey
	}

	file, err := ioutil.TempFile("", "import")
	if err != nil {
		return nil, nil, xerrors.Errorf("could not create temporary file: %w", err)
	}
	defer func() {
		os.Remove(file.Name())
		file.Close()
	}()
	if _, err := file.Write(key); err != nil {
		return nil, nil, xerrors.Errorf("could not write key to temporary file: %w", err)
	}
	cmd := exec.Command("gpg", "--import", "--armor")
	cmd.Args = append(cmd.Args, g.options()...)
	cmd.Args = append(cmd.Args, file.Name())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, nil, xerrors.Errorf("could not import new key (stderr: %q): %w", stderr.String(), err)
	}
	if err := g.Refresh(); err != nil {
		return nil, nil, xerrors.Errorf("could not refresh after import (keyring is inconsistent): %w", err)
	}

	pkeys := g.PublicKeys()
	var newPublicKeys []*PublicKey
	for _, pkey := range pkeys {
		if _, ok := oldPublicKeys[pkey.Fingerprint]; !ok {
			newPublicKeys = append(newPublicKeys, pkey)
		}
	}
	skeys := g.SecretKeys()
	var newSecretKeys []*SecretKey
	for _, skey := range skeys {
		if _, ok := oldSecretKeys[skey.Fingerprint]; !ok {
			newSecretKeys = append(newSecretKeys, skey)
		}
	}
	return newPublicKeys, newSecretKeys, nil
}

func (g *Keyring) SetTrusted(fingerprint string, trusted bool) error {
	cmd := exec.Command("gpg", "--command-fd", "0", "--edit-key", fingerprint)
	cmd.Args = append(cmd.Args, g.options()...)
	if trusted {
		cmd.Stdin = strings.NewReader("trust\n5\ny\n")
	} else {
		cmd.Stdin = strings.NewReader("trust\n2\n")
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		log.Printf("error: update key trust cmd %v failed with stderr %q", cmd.Args, stderr.String())
		return xerrors.Errorf("could not update trust status on key %q: %w", fingerprint, err)
	}
	log.Printf("trust status for key %q is updated to %t (stdout %q)", fingerprint, trusted, stdout.String())
	return nil
}

type PublicKey struct {
	KeyID       string `json:"key_id"`
	Fingerprint string `json:"fingerprint"`
	UserHash    string `json:"user_hash"`

	CanEncrypt bool `json:"can_encrypt"`

	Trusted bool `json:"trusted"`
	Subkey  bool `json:"subkey"`

	KeyLength int       `json:"key_length"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`

	UserDetail string `json:"user_detail"`
	UserName   string `json:"user_name"`
	UserEmail  string `json:"user_email"`
}

type SecretKey struct {
	KeyID       string `json:"key_id"`
	Fingerprint string `json:"fingerprint"`
	UserHash    string `json:"user_hash"`

	CanEncrypt bool `json:"can_encrypt"`

	Trusted bool `json:"trusted"`
	Subkey  bool `json:"subkey"`

	KeyLength int       `json:"key_length"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`

	UserDetail string `json:"user_detail"`
	UserName   string `json:"user_name"`
	UserEmail  string `json:"user_email"`
}

func (g *Keyring) SecretKeys() []*SecretKey {
	var groups [][]*internal.Record
	var recs []*internal.Record
	for _, rec := range g.skeyRecords {
		if rec.IsSecretKey() {
			if len(recs) > 0 {
				groups = append(groups, recs)
			}
			recs = []*internal.Record{rec}
			continue
		}
		if len(recs) > 0 {
			recs = append(recs, rec)
		}
	}
	if len(recs) > 0 {
		groups = append(groups, recs)
	}

	var sks []*SecretKey
	for _, recs := range groups {
		skey, sskeys := g.parseSecretKey(recs)
		if skey != nil {
			sks = append(sks, skey)
		}
		if sskeys != nil {
			sks = append(sks, sskeys...)
		}
	}
	return sks
}

func (g *Keyring) parseSecretKey(recs []*internal.Record) (*SecretKey, []*SecretKey) {
	if !recs[0].IsSecretKey() {
		return nil, nil
	}
	sec := &SecretKey{}
	ssubs := []*SecretKey{}
	for ii, rec := range recs {
		if rec.IsSecretSubkey() {
			ssubs = g.parseSecretSubkeys(sec, recs[ii:])
			break
		} else if rec.IsSecretKey() {
			sec.KeyID = rec.KeyID()
			sec.KeyLength = rec.KeyLength()
			sec.CanEncrypt = rec.CanEncrypt()
			sec.CreatedAt = rec.CreatedAt()
			sec.ExpiresAt = rec.ExpiresAt()
			sec.Trusted = rec.KeyTrusted()
		} else if rec.IsUserID() {
			sec.UserDetail = rec.UserDetail()
			sec.UserEmail = rec.UserEmail()
			sec.UserHash = rec.UserHash()
			sec.UserName = rec.UserName()
		} else if rec.IsFingerprint() {
			sec.Fingerprint = rec.Fingerprint()
		}
	}
	return sec, ssubs
}

func (g *Keyring) parseSecretSubkeys(sec *SecretKey, recs []*internal.Record) []*SecretKey {
	if !recs[0].IsSecretSubkey() {
		return nil
	}
	var sub *SecretKey
	var subs []*SecretKey
	for _, rec := range recs {
		if !rec.IsSecretSubkey() && !rec.IsFingerprint() {
			break
		}
		if rec.IsSecretSubkey() {
			if sub != nil {
				subs = append(subs, sub)
			}
			sub = &SecretKey{
				Subkey:     true,
				UserDetail: sec.UserDetail,
				UserEmail:  sec.UserEmail,
				UserHash:   sec.UserHash,
				UserName:   sec.UserName,
				KeyID:      rec.KeyID(),
				KeyLength:  rec.KeyLength(),
				CanEncrypt: rec.CanEncrypt(),
				CreatedAt:  rec.CreatedAt(),
				ExpiresAt:  rec.ExpiresAt(),
				Trusted:    rec.KeyTrusted(),
			}
		}
		if rec.IsFingerprint() {
			sub.Fingerprint = rec.Fingerprint()
		}
	}
	if sub != nil {
		subs = append(subs, sub)
	}
	return subs
}

func (g *Keyring) PublicKeys() []*PublicKey {
	var groups [][]*internal.Record
	var recs []*internal.Record
	for _, rec := range g.keyRecords {
		if rec.IsPublicKey() {
			if len(recs) > 0 {
				groups = append(groups, recs)
			}
			recs = []*internal.Record{rec}
			continue
		}
		if len(recs) > 0 {
			recs = append(recs, rec)
		}
	}
	if len(recs) > 0 {
		groups = append(groups, recs)
	}

	var pks []*PublicKey
	for _, recs := range groups {
		pkey, skeys := g.parsePublicKeys(recs)
		if pkey != nil {
			pks = append(pks, pkey)
		}
		if skeys != nil {
			pks = append(pks, skeys...)
		}
	}
	return pks
}

func (g *Keyring) parsePublicKeys(recs []*internal.Record) (*PublicKey, []*PublicKey) {
	if !recs[0].IsPublicKey() {
		return nil, nil
	}
	pub := &PublicKey{}
	subs := []*PublicKey{}
	for ii, rec := range recs {
		if rec.IsSubkey() {
			subs = g.parseSubkeys(pub, recs[ii:])
			break
		} else if rec.IsPublicKey() {
			pub.KeyID = rec.KeyID()
			pub.KeyLength = rec.KeyLength()
			pub.CanEncrypt = rec.CanEncrypt()
			pub.CreatedAt = rec.CreatedAt()
			pub.ExpiresAt = rec.ExpiresAt()
			pub.Trusted = rec.KeyTrusted()
		} else if rec.IsUserID() {
			pub.UserDetail = rec.UserDetail()
			pub.UserEmail = rec.UserEmail()
			pub.UserHash = rec.UserHash()
			pub.UserName = rec.UserName()
		} else if rec.IsFingerprint() {
			pub.Fingerprint = rec.Fingerprint()
		}
	}
	return pub, subs
}

func (g *Keyring) parseSubkeys(pub *PublicKey, recs []*internal.Record) []*PublicKey {
	if !recs[0].IsSubkey() {
		return nil
	}
	var sub *PublicKey
	var subs []*PublicKey
	for _, rec := range recs {
		if !rec.IsSubkey() && !rec.IsFingerprint() {
			break
		}
		if rec.IsSubkey() {
			if sub != nil {
				subs = append(subs, sub)
			}
			sub = &PublicKey{
				Subkey:     true,
				UserDetail: pub.UserDetail,
				UserEmail:  pub.UserEmail,
				UserHash:   pub.UserHash,
				UserName:   pub.UserName,
				KeyID:      rec.KeyID(),
				KeyLength:  rec.KeyLength(),
				CanEncrypt: rec.CanEncrypt(),
				CreatedAt:  rec.CreatedAt(),
				ExpiresAt:  rec.ExpiresAt(),
				Trusted:    rec.KeyTrusted(),
			}
		}
		if rec.IsFingerprint() {
			sub.Fingerprint = rec.Fingerprint()
		}
	}
	if sub != nil {
		subs = append(subs, sub)
	}
	return subs
}

func (g *Keyring) ListPublicKeyFingerprints() []string {
	var fps []string
	for ii, rec := range g.keyRecords {
		if rec.IsPublicKey() {
			for jj := ii + 1; jj < len(g.keyRecords); jj++ {
				rec := g.keyRecords[jj]
				if rec.IsFingerprint() {
					fps = append(fps, rec.Fingerprint())
					break
				}
			}
		}
	}
	return fps
}

func (g *Keyring) ListEncryptionKeys() []string {
	var fps []string
	for ii, rec := range g.keyRecords {
		if !rec.IsPublicKey() || !rec.CanEncrypt() {
			continue
		}
		for jj := ii + 1; jj < len(g.keyRecords); jj++ {
			rec := g.keyRecords[jj]
			if rec.IsFingerprint() {
				fps = append(fps, rec.Fingerprint())
				break
			}
		}
	}
	return fps
}

func (g *Keyring) UserHashForEmail(email string) []string {
	var uhs []string
	for _, rec := range g.keyRecords {
		if rec.IsUserID() {
			detail := rec.UserDetail()
			if strings.HasSuffix(detail, "<"+email+">") {
				uhs = append(uhs, rec.UserHash())
			}
		}
	}
	return uhs
}

func (g *Keyring) parseRecords(output string) ([]*internal.Record, error) {
	var records []*internal.Record
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Split(strings.TrimSpace(line), ":")
		records = append(records, internal.NewRecord("", fields))
	}
	return records, nil
}

func GetRecipients(file string) ([]string, error) {
	cmd := exec.Command("gpg", "--list-only", "-d", file)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		log.Printf("get recipients cmd %v failed with stderr %q", cmd.Args, stderr.String())
		return nil, xerrors.Errorf("could not determine recipients: %w", err)
	}
	trim := func(r rune) bool {
		return !strings.ContainsRune("0123456789ABCDEFabcdef", r)
	}
	// For some reason, GPG writes the ids to stderr.
	var ids []string
	fields := strings.Fields(stderr.String())
	for ii := 1; ii < len(fields); ii++ {
		if fields[ii-1] == "ID" {
			ids = append(ids, strings.TrimFunc(fields[ii], trim))
		}
	}
	return ids, nil
}
