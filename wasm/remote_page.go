// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"golang.org/x/xerrors"
	"honnef.co/go/js/dom/v2"

	"github.com/bvk/past/msg"
)

const remoteTemplate = `
<div>
	<div class="row header">
		<button class="column material-icons" gotag="name:BackButton click:OnBackButton">navigate_before</button>
		<span class="column-elastic header-title">Remote Repository Status</span>
		<button class="column material-icons" gotag="name:CloseButton click:OnCloseButton">clear</button>
	</div>

	<hr/>

	<div class="content mw32em">
		<div class="row">
			<span class="column-lefty bold">Local Tip</span>
		</div>

		<div class="row">
			<span class="column-righty" gotag="name:LocalAuthorDate"></span>
		</div>
		<div class="row">
			<span class="column-righty" gotag="name:LocalAuthor"></span>
		</div>
		<div class="row">
			<span class="column-righty" gotag="name:LocalCommit"></span>
		</div>
		<div class="row">
			<span class="column-elastic" gotag="name:LocalTitle"></span>
		</div>

		<div class="row">
			<span class="column-lefty bold">Remote Tip</span>
			<button class="column material-icons" gotag="name:FetchButton click:OnFetchButton">refresh</button>
		</div>

		<div class="row">
			<span class="column-righty text-right" gotag="name:RemoteAuthorDate"></span>
		</div>
		<div class="row">
			<span class="column-righty text-right" gotag="name:RemoteAuthor"></span>
		</div>
		<div class="row">
			<span class="column-righty text-right" gotag="name:RemoteCommit"></span>
		</div>
		<div class="row">
			<span class="column-elastic" gotag="name:RemoteTitle"></span>
		</div>
	</div>

	<div class="row footer">
		<button class="column material-icons" gotag="name:PushButton click:OnPushButton" disabled>cloud_upload</button>
		<div class="column-elastic footer-status"></div>
		<button class="column material-icons" gotag="name:PullButton click:OnPullButton" disabled>cloud_download</button>
	</div>
</div>
`

type RemotePage struct {
	ctl *Controller

	data *msg.SyncRemoteResponse

	root dom.Element

	BackButton, CloseButton *dom.HTMLButtonElement

	PushButton, PullButton, FetchButton *dom.HTMLButtonElement

	LocalAuthorDate, LocalAuthor, LocalCommit, LocalTitle *dom.HTMLSpanElement

	RemoteAuthorDate, RemoteAuthor, RemoteCommit, RemoteTitle *dom.HTMLSpanElement
}

func NewRemotePage(ctl *Controller, params *msg.SyncRemoteResponse) (*RemotePage, error) {
	p := &RemotePage{
		ctl:  ctl,
		data: params,
	}
	root, err := BindHTML(remoteTemplate, "gotag", p)
	if err != nil {
		return nil, xerrors.Errorf("could not parse remote page template: %w", err)
	}
	p.root = root
	return p, nil
}

func (p *RemotePage) RootDiv() dom.Element {
	return p.root
}

func (p *RemotePage) RefreshDisplay() error {
	p.LocalCommit.SetTextContent(p.data.Head.Commit)
	p.LocalAuthor.SetTextContent(p.data.Head.Author)
	p.LocalAuthorDate.SetTextContent(p.data.Head.AuthorDate.String())
	p.LocalTitle.SetTextContent(p.data.Head.Title)

	p.RemoteCommit.SetTextContent(p.data.Head.Commit)
	p.RemoteAuthor.SetTextContent(p.data.Head.Author)
	p.RemoteAuthorDate.SetTextContent(p.data.Head.AuthorDate.String())
	p.RemoteTitle.SetTextContent(p.data.Head.Title)

	p.PushButton.SetTextContent(PushRemoteIcon)
	p.PullButton.SetTextContent(PullRemoteIcon)
	if p.data.Head.Commit == p.data.Remote.Commit {
		p.PullButton.SetDisabled(true)
		p.PushButton.SetDisabled(true)
		return nil
	}

	if p.data.NewerCommit == p.data.Remote.Commit {
		p.PullButton.SetDisabled(false)
		p.PushButton.SetDisabled(true)
		return nil
	}

	if p.data.NewerCommit == p.data.Head.Commit {
		p.PullButton.SetDisabled(true)
		p.PushButton.SetDisabled(false)
		return nil
	}

	p.PushButton.SetDisabled(false)
	p.PullButton.SetDisabled(false)
	p.PushButton.SetTextContent(UploadIcon)
	p.PullButton.SetTextContent(DownloadIcon)
	return nil
}

func (p *RemotePage) OnBackButton(dom.Event) (status error) {
	resp, err := p.ctl.CheckStatus()
	if err != nil {
		return p.ctl.setError(err)
	}
	page, err := NewSettingsPage(p.ctl, resp)
	if err != nil {
		return p.ctl.setError(err)
	}
	if err := p.ctl.ShowPage(page); err != nil {
		return err
	}
	return nil
}

func (p *RemotePage) OnCloseButton(dom.Event) (status error) {
	return p.ctl.ClosePage(p)
}

func (p *RemotePage) OnFetchButton(dom.Event) (status error) {
	resp, err := p.ctl.SyncRemote()
	if err != nil {
		return p.ctl.setError(err)
	}
	p.data = resp
	return p.RefreshDisplay()
}

func (p *RemotePage) OnPushButton(dom.Event) (status error) {
	resp, err := p.ctl.PushRemote()
	if err != nil {
		return p.ctl.setError(err)
	}
	p.data = resp
	return p.RefreshDisplay()
}

func (p *RemotePage) OnPullButton(dom.Event) (status error) {
	resp, err := p.ctl.PullRemote()
	if err != nil {
		return p.ctl.setError(err)
	}
	p.data = resp
	return p.RefreshDisplay()
}
