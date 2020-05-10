// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"fmt"

	"golang.org/x/xerrors"
	"honnef.co/go/js/dom/v2"

	"github.com/bvk/past"
	"github.com/bvk/past/msg"
)

const pastTemplate = `
<div>
	<div class="row header">
		<button class="cell material-icons" gotag="name:BackButton click:OnBackButton">navigate_before</button>
		<span class="cell-elastic header-title">Password Store Status</span>
		<button class="cell material-icons" gotag="name:CloseButton click:OnCloseButton">clear</button>
	</div>

	<hr/>

	<div class="content mw32em">
		<div class="row">
			<span class="cell-lefty nowrap bold">Unknown Recipients</span>
		</div>

		<ul>
			<li gotag="name:MissingKeyTemplate" style="display:none">
				<div class="row">
					<span class="cell-lefty" missing="name:KeyID"></span>
					<span class="cell-righty">Can read</span>
					<span class="cell" missing="name:FileCount"></span>
					<span class="cell">files</span>
				</div>
			</li>
		</ul>

		<div class="row">
			<span class="cell-lefty nowrap bold">Recipients in Use</span>
		</div>

		<ul>
			<li gotag="name:RecipientTemplate" style="display:none">
				<div class="row">
					<span class="cell material-icons">360</span>
					<span class="cell-lefty" recipient="name:KeyID click:OnKeyID"></span>
					<span class="cell" recipient="name:ReadCount"></span>
					<span class="cell" recipient="name:Reason" style="display:none"></span>
					<button class="cell material-icons" recipient="name:RemoveButton click:OnRemove">delete</button>
				</div>
			</li>
		</ul>

		<div class="row">
			<span class="cell-lefty nowrap bold">Available Recipients</span>
		</div>

		<ul>
			<li gotag="name:AvailableTemplate" style="display:none">
				<div class="row">
					<span class="cell material-icons">360</span>
					<span class="cell-lefty" available="name:KeyID click:OnKeyID"></span>
					<span class="cell disabled" available="name:Reason" style="display:none"></span>
					<button class="cell material-icons" available="name:AddButton click:OnAdd">add</button>
				</div>
			</li>
		</ul>

	</div>

	<div class="row footer">
		<div class="cell-elastic footer-status"></div>
	</div>
</div>
`

type PastPage struct {
	ctl *Controller

	data *msg.ScanStoreResponse

	root dom.Element

	BackButton, CloseButton *dom.HTMLButtonElement

	MissingKeyTemplate, RecipientTemplate, AvailableTemplate *dom.HTMLLIElement
}

func NewPastPage(ctl *Controller, params *msg.ScanStoreResponse) (*PastPage, error) {
	p := &PastPage{
		ctl:  ctl,
		data: params,
	}
	root, err := BindHTML(pastTemplate, "gotag", p)
	if err != nil {
		return nil, xerrors.Errorf("could not parse password store page template: %w", err)
	}
	p.root = root
	return p, nil
}

func (p *PastPage) RootDiv() dom.Element {
	return p.root
}

func (p *PastPage) RefreshDisplay() error {
	// Remove dynamic elements if any were added previously.
	for n := p.MissingKeyTemplate.NextSibling(); n != nil; n = p.MissingKeyTemplate.NextSibling() {
		p.MissingKeyTemplate.ParentNode().RemoveChild(n)
	}
	for n := p.RecipientTemplate.NextSibling(); n != nil; n = p.RecipientTemplate.NextSibling() {
		p.RecipientTemplate.ParentNode().RemoveChild(n)
	}
	for n := p.AvailableTemplate.NextSibling(); n != nil; n = p.AvailableTemplate.NextSibling() {
		p.AvailableTemplate.ParentNode().RemoveChild(n)
	}

	for keyid, count := range p.data.MissingKeyFileCountMap {
		clone := mustCloneHTMLElement(p.MissingKeyTemplate, true)

		k := new(struct{ KeyID, FileCount *dom.HTMLSpanElement })
		if err := BindTarget(clone, "missing", k); err != nil {
			return p.ctl.setError(err)
		}
		k.KeyID.SetTextContent(keyid)
		k.FileCount.SetTextContent(fmt.Sprintf("%d", count))

		clone.Style().RemoveProperty("display")
		p.MissingKeyTemplate.ParentNode().InsertBefore(clone, p.MissingKeyTemplate.NextSibling())
	}

	for fp, count := range p.data.KeyFileCountMap {
		clone := mustCloneHTMLElement(p.RecipientTemplate, true)

		key := p.data.KeyMap[fp]
		k := &pastStatusKeyListItem{page: p, key: key}
		if err := BindTarget(clone, "recipient", k); err != nil {
			return p.ctl.setError(err)
		}
		k.KeyID.SetTextContent(key.UserName)
		k.ReadCount.SetTextContent(fmt.Sprintf("%d/%d", count, p.data.NumFiles))

		clone.Style().RemoveProperty("display")
		p.RecipientTemplate.ParentNode().InsertBefore(clone, p.RecipientTemplate.NextSibling())
	}

	for _, key := range p.data.UnusedKeyMap {
		clone := mustCloneHTMLElement(p.AvailableTemplate, true)

		k := &pastStatusKeyListItem{page: p, key: key}
		if err := BindTarget(clone, "available", k); err != nil {
			return p.ctl.setError(err)
		}
		k.KeyID.SetTextContent(key.UserName)
		if !key.IsTrusted {
			k.Reason.SetTextContent("untrusted")
			k.Reason.Style().RemoveProperty("display")
			k.AddButton.SetDisabled(true)
		} else if key.IsExpired {
			k.Reason.SetTextContent("expired")
			k.Reason.Style().RemoveProperty("display")
			k.AddButton.SetDisabled(true)
		}

		clone.Style().RemoveProperty("display")
		p.AvailableTemplate.ParentNode().InsertBefore(clone, p.AvailableTemplate.NextSibling())
	}
	return nil
}

func (p *PastPage) OnBackButton(dom.Event) (status error) {
	if _, err := ShowSettingsPage(p.ctl); err != nil {
		return p.ctl.setError(err)
	}
	return nil
}

func (p *PastPage) OnCloseButton(dom.Event) (status error) {
	return p.ctl.ClosePage(p)
}

func (p *PastPage) numReadable() int {
	n := 0
	for fp, count := range p.data.KeyFileCountMap {
		key := p.data.KeyMap[fp]
		if key.CanDecrypt && count > n {
			n = count
		}
	}
	return n
}

type pastStatusKeyListItem struct {
	page *PastPage

	key *past.PublicKeyData

	KeyID, ReadCount, Reason *dom.HTMLSpanElement

	AddButton, RemoveButton *dom.HTMLButtonElement
}

func (k *pastStatusKeyListItem) OnKeyID(dom.Event) (status error) {
	rotateKeyID(k.KeyID, k.key)
	return nil
}

func (k *pastStatusKeyListItem) OnRemove(dom.Event) (status error) {
	nreadable := k.page.numReadable()
	req := &msg.RemoveRecipientRequest{
		ScanStore:   true,
		Fingerprint: k.key.KeyFingerprint,
		NumSkip:     k.page.data.NumFiles - nreadable,
	}

	bs := StartBusy(k.page.root)
	resp, err := k.page.ctl.RemoveRecipient(req)
	StopBusy(bs)

	if err != nil {
		return k.page.ctl.setError(err)
	}
	k.page.data = resp.ScanStore
	return k.page.RefreshDisplay()
}

func (k *pastStatusKeyListItem) OnAdd(dom.Event) (status error) {
	nreadable := k.page.numReadable()
	req := &msg.AddRecipientRequest{
		ScanStore:   true,
		Fingerprint: k.key.KeyFingerprint,
		NumSkip:     k.page.data.NumFiles - nreadable,
	}

	bs := StartBusy(k.page.root)
	resp, err := k.page.ctl.AddRecipient(req)
	StopBusy(bs)

	if err != nil {
		return k.page.ctl.setError(err)
	}
	k.page.data = resp.ScanStore
	return k.page.RefreshDisplay()
}
