// Copyright (c) 2020 BVK Chaitanya

package gpg

import (
	"bytes"
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

func Create(name, email, passphrase string) (*Keyring, error) {
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

	content := `
Key-Type: default
Key-Length: 4096
Subkey-Type: default
Subkey-Length: 4096
Name-Real: ` + name + `
Name-Email: ` + email + `
Expire-Date: 0
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
		"--quiet",
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

func (g *Keyring) Encrypt(data []byte, reps []*PublicKey) ([]byte, error) {
	opts := []string{
		"--compress-algo=none",
	}
	for _, rep := range reps {
		opts = append(opts, "--recipient", rep.Fingerprint)
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
