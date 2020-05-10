// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"golang.org/x/xerrors"
	"honnef.co/go/js/dom/v2"

	"github.com/bvk/past"
	"github.com/bvk/past/msg"
)

const keyringTemplate = `
<div>
	<div class="row header">
		<button class="cell material-icons" gotag="name:BackButton click:OnBackButton">navigate_before</button>
		<span class="cell-elastic header-title">GPG Keyring</span>
		<button class="cell material-icons" gotag="name:CloseButton click:OnCloseButton">clear</button>
	</div>

  <hr />

	<div class="content mw32em">
		<div class="row">
			<span class="cell-elastic bold nowrap">Local Keys (with Private Key)</span>
		</div>

		<ul>
			<li gotag="name:LocalKeyTemplate" style="display:none">
				<div class="row">
					<span class="cell material-icons">360</span>
					<span class="cell-lefty" localkey="name:KeyID click:OnKeyID"></span>
					<button class="cell material-icons" localkey="name:ViewKey click:OnViewKey">navigate_next</button>
				</div>
			</li>
		</ul>

		<div class="row">
			<span class="cell-elastic bold nowrap">Remote Keys (without Private Key)</span>
		</div>

		<ul>
			<li gotag="name:RemoteKeyTemplate" style="display:none">
				<div class="row">
					<span class="cell material-icons">360</span>
					<span class="cell-lefty" remotekey="name:KeyID click:OnKeyID"></span>
					<button class="cell material-icons" remotekey="name:ViewKey click:OnViewKey">navigate_next</button>
				</div>
			</li>
		</ul>

		<div class="row">
			<span class="cell-elastic bold nowrap">Expired Keys</span>
		</div>

		<ul>
			<li gotag="name:ExpiredKeyTemplate" style="display:none">
				<div class="row">
					<span class="cell material-icons">360</span>
					<span class="cell-lefty" expiredkey="name:KeyID click:OnKeyID"></span>
					<button class="cell material-icons" expiredkey="name:ViewKey click:OnViewKey">navigate_next</button>
				</div>
			</li>
		</ul>
	</div>

	<div class="row footer">
		<button class="cell material-icons" gotag="name:AddButton click:OnAddButton">add</button>
		<div class="cell-elastic footer-status"></div>
		<button class="cell material-icons" gotag="name:AddButton2 click:OnAddButton">add</button>
	</div>
</div>
`

type KeyringPage struct {
	ctl *Controller

	args *msg.CheckStatusResponse

	root dom.Element

	BackButton, CloseButton *dom.HTMLButtonElement

	AddButton, AddButton2 *dom.HTMLButtonElement

	LocalKeyTemplate, RemoteKeyTemplate, ExpiredKeyTemplate *dom.HTMLLIElement
}

func NewKeyringPage(ctl *Controller, params *msg.CheckStatusResponse) (*KeyringPage, error) {
	p := new(KeyringPage)
	root, err := BindHTML(keyringTemplate, "gotag", p)
	if err != nil {
		return nil, xerrors.Errorf("could not parse keyring page template: %w", err)
	}
	p.ctl, p.args, p.root = ctl, params, root
	return p, nil
}

func (p *KeyringPage) RootDiv() dom.Element {
	return p.root
}

func (p *KeyringPage) RefreshDisplay() error {
	// Remove dynamic elements if any were added previously.
	for n := p.LocalKeyTemplate.NextSibling(); n != nil; n = p.LocalKeyTemplate.NextSibling() {
		p.LocalKeyTemplate.ParentNode().RemoveChild(n)
	}
	for n := p.RemoteKeyTemplate.NextSibling(); n != nil; n = p.RemoteKeyTemplate.NextSibling() {
		p.RemoteKeyTemplate.ParentNode().RemoveChild(n)
	}
	for n := p.ExpiredKeyTemplate.NextSibling(); n != nil; n = p.ExpiredKeyTemplate.NextSibling() {
		p.ExpiredKeyTemplate.ParentNode().RemoveChild(n)
	}

	for _, key := range p.args.LocalKeys {
		clone := mustCloneHTMLElement(p.LocalKeyTemplate, true)

		key := &KeyListItem{ctl: p.ctl, key: key}
		if err := BindTarget(clone, "localkey", key); err != nil {
			return p.ctl.setError(err)
		}
		key.KeyID.SetTextContent(key.key.UserName)

		clone.Style().RemoveProperty("display")
		p.LocalKeyTemplate.ParentNode().InsertBefore(clone, p.LocalKeyTemplate.NextSibling())
	}
	for _, key := range p.args.RemoteKeys {
		clone := mustCloneHTMLElement(p.RemoteKeyTemplate, true)

		key := &KeyListItem{ctl: p.ctl, key: key}
		if err := BindTarget(clone, "remotekey", key); err != nil {
			return p.ctl.setError(err)
		}
		key.KeyID.SetTextContent(key.key.UserName)

		clone.Style().RemoveProperty("display")
		p.RemoteKeyTemplate.ParentNode().InsertBefore(clone, p.RemoteKeyTemplate.NextSibling())
	}
	for _, key := range p.args.ExpiredKeys {
		clone := mustCloneHTMLElement(p.ExpiredKeyTemplate, true)

		key := &KeyListItem{ctl: p.ctl, key: key}
		if err := BindTarget(clone, "expiredkey", key); err != nil {
			return p.ctl.setError(err)
		}
		key.KeyID.SetTextContent(key.key.UserName)

		clone.Style().RemoveProperty("display")
		p.ExpiredKeyTemplate.ParentNode().InsertBefore(clone, p.ExpiredKeyTemplate.NextSibling())
	}
	return nil
}

func (p *KeyringPage) OnBackButton(dom.Event) (status error) {
	if _, err := ShowSettingsPage(p.ctl); err != nil {
		return p.ctl.setError(err)
	}
	return nil
}

func (p *KeyringPage) OnCloseButton(dom.Event) (status error) {
	return p.ctl.ClosePage(p)
}

func (p *KeyringPage) OnAddButton(dom.Event) (status error) {
	page, err := NewAddKeyPage(p.ctl)
	if err != nil {
		return p.ctl.setError(err)
	}
	if err := p.ctl.ShowPage(page); err != nil {
		return err
	}
	return nil
}

type KeyListItem struct {
	ctl *Controller

	key *past.PublicKeyData

	KeyID *dom.HTMLSpanElement

	ViewKey *dom.HTMLButtonElement
}

func (k *KeyListItem) OnKeyID(dom.Event) (status error) {
	rotateKeyID(k.KeyID, k.key)
	return nil
}

func (k *KeyListItem) OnViewKey(dom.Event) (status error) {
	page, err := NewKeyPage(k.ctl, k.key)
	if err != nil {
		return k.ctl.setError(err)
	}
	if err := k.ctl.ShowPage(page); err != nil {
		return err
	}
	return nil
}
