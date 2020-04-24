// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"fmt"
	"os"

	"golang.org/x/xerrors"
	"honnef.co/go/js/dom/v2"

	"github.com/bvk/past"
	"github.com/bvk/past/msg"
)

const keyTemplate = `
<div>
	<div class="row header">
		<button class="column material-icons" gotag="name:BackButton click:OnBackButton">navigate_before</button>
		<span class="column-elastic header-title">GPG Key Details</span>
		<button class="column material-icons" gotag="name:CloseButton click:OnCloseButton">clear</button>
	</div>

	<hr/>

	<div class="content mw32em">
		<div class="row">
			<span class="column-lefty" gotag="name:KeyFingerprint"></span>
			<button class="column material-icons" gotag="name:CopyButton click:OnCopyButton">content_copy</button>
		</div>

		<div class="row">
			<span class="column-lefty" gotag="name:KeyUserName"></span>
			<span class="column-righty" gotag="name:KeyTrusted"></span>
			<button class="column material-icons" gotag="name:ToggleTrust click:OnToggleTrust">check_circle_outline</button>
		</div>

		<div class="row">
			<span class="column-lefty" gotag="name:KeyUserEmail"></span>
			<span class="column-righty" gotag="name:KeyExpired"></span>
		</div>
	</div>

	<div class="row footer">
		<button class="column material-icons" gotag="name:ExportButton click:OnExportButton">file_copy</button>
		<div class="column-elastic footer-status"></div>
		<button class="column material-icons" gotag="name:DeleteButton click:OnDeleteButton">delete</button>
	</div>
</div>
`

type KeyPage struct {
	ctl *Controller

	key *past.PublicKeyData

	root dom.Element

	BackButton, CloseButton    *dom.HTMLButtonElement
	ExportButton, DeleteButton *dom.HTMLButtonElement

	CopyButton, ToggleTrust *dom.HTMLButtonElement

	KeyFingerprint, KeyUserName, KeyUserEmail *dom.HTMLSpanElement

	KeyExpired, KeyTrusted *dom.HTMLSpanElement
}

func NewKeyPage(ctl *Controller, key *past.PublicKeyData) (*KeyPage, error) {
	p := &KeyPage{
		ctl: ctl,
		key: key,
	}
	root, err := BindHTML(keyTemplate, "gotag", p)
	if err != nil {
		return nil, xerrors.Errorf("could not parse key page template: %w", err)
	}
	p.root = root
	return p, nil
}

func (p *KeyPage) RootDiv() dom.Element {
	return p.root
}

func (p *KeyPage) RefreshDisplay() error {
	// TODO: Remove key- prefix to these fields.
	p.KeyFingerprint.SetTextContent(p.key.KeyFingerprint)
	p.KeyUserName.SetTextContent(p.key.UserName)
	p.KeyUserEmail.SetTextContent(p.key.UserEmail)
	if p.key.IsTrusted {
		p.KeyTrusted.SetTextContent("Trusted")
	} else {
		p.KeyTrusted.SetTextContent("Not-Trusted")
	}

	if p.key.IsTrusted {
		p.ToggleTrust.SetTextContent(UntrustIcon)
	} else {
		p.ToggleTrust.SetTextContent(TrustIcon)
	}

	if p.key.IsExpired {
		p.ToggleTrust.SetDisabled(true)
		p.KeyExpired.SetTextContent("Expired")
	} else if p.key.DaysToExpire == 0 {
		p.KeyExpired.SetTextContent("Never Expires")
	} else {
		p.KeyExpired.SetTextContent(fmt.Sprintf("Expires in %d Days", p.key.DaysToExpire))
	}
	return nil
}

func (p *KeyPage) OnBackButton(dom.Event) (status error) {
	resp, err := p.ctl.CheckStatus()
	if err != nil {
		return p.ctl.setError(err)
	}
	page, err := NewKeyringPage(p.ctl, resp.CheckStatus)
	if err != nil {
		return p.ctl.setError(err)
	}
	if err := p.ctl.ShowPage(page); err != nil {
		return err
	}
	return nil
}

func (p *KeyPage) OnCloseButton(dom.Event) (status error) {
	return p.ctl.ClosePage(p)
}

func (p *KeyPage) OnToggleTrust(dom.Event) (status error) {
	req := &msg.EditKeyRequest{
		Fingerprint: p.key.KeyFingerprint,
		Trust:       !p.key.IsTrusted,
	}
	resp, err := p.ctl.EditKey(req)
	if err != nil {
		return p.ctl.setError(err)
	}
	p.key = resp.Key
	return p.RefreshDisplay()
}

func (p *KeyPage) OnCopyButton(dom.Event) (status error) {
	if err := p.ctl.Copy(p.key.KeyFingerprint); err != nil {
		return p.ctl.setError(err)
	}
	p.ctl.setGoodStatus("Copied")
	return nil
}

func (p *KeyPage) OnExportButton(dom.Event) (status error) {
	req := &msg.ExportKeyRequest{
		Fingerprint: p.key.KeyFingerprint,
	}
	resp, err := p.ctl.ExportKey(req)
	if err != nil {
		return p.ctl.setError(err)
	}
	if err := p.ctl.Copy(resp.ArmorKey); err != nil {
		p.ctl.setBadStatus("Could not copy")
		return os.ErrInvalid
	}
	p.ctl.setGoodStatus("Copied")
	return nil
}

func (p *KeyPage) OnDeleteButton(dom.Event) (status error) {
	req := &msg.DeleteKeyRequest{
		CheckStatus: true,
		Fingerprint: p.key.KeyFingerprint,
	}
	resp, err := p.ctl.DeleteKey(req)
	if err != nil {
		return p.ctl.setError(err)
	}
	page, err := NewKeyringPage(p.ctl, resp.CheckStatus)
	if err != nil {
		return p.ctl.setError(err)
	}
	if err := p.ctl.ShowPage(page); err != nil {
		return err
	}
	return nil
}
