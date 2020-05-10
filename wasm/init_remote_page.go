// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"strings"

	"golang.org/x/xerrors"
	"honnef.co/go/js/dom/v2"

	"github.com/bvk/past/msg"
)

const initRemoteTemplate = `
<div>
	<div class="row header">
		<button class="cell material-icons" gotag="name:BackButton click:OnBackButton">navigate_before</button>
		<span class="cell-elastic header-title">Setup Remote Repository</span>
		<button class="cell material-icons" gotag="name:CloseButton click:OnCloseButton">clear</button>
	</div>

	<hr/>

	<div class="content mw32em">
		<div class="row">
			<span class="cell w6em">Provider</span>
			<select class="cell-elastic" gotag="name:GitServer change:RefreshButtons">
				<optgroup>
					<option value="ssh">Custom Server using SSH protocol</option>
					<option value="https">Custom Server using HTTPS protocol</option>
					<option value="git">Custom Server using GIT protocol</option>
				</optgroup>
				<optgroup>
					<option value="github-ssh">Github using SSH protocol</option>
					<option value="github-https">Github using HTTPS protocol</option>
				</optgroup>
			</select>
		</div>

		<div class="row">
			<span class="cell w6em">Username</span>
			<input class="cell-elastic" gotag="name:GitUser input:RefreshButtons" placeholder="username"></input>
		</div>

		<div class="row">
			<span class="cell w6em">Password</span>
			<input class="cell-elastic" gotag="name:GitPass input:RefreshButtons" type="password" placeholder="leave empty for password-less authentication"></input>
			<button class="cell material-icons" gotag="name:ToggleGitPassButton click:OnToggleGitPass">visibility_off</button>
		</div>

		<div class="row">
			<span class="cell w6em">Hostname</span>
			<input class="cell-elastic" gotag="name:GitHost input:RefreshButtons" placeholder="hostname.com:1234"></input>
		</div>

		<div class="row">
			<span class="cell w6em">Repo Path</span>
			<input class="cell-elastic" gotag="name:GitPath input:RefreshButtons" placeholder="path/to/repository.git"></input>
		</div>
	</div>

	<div class="row footer">
		<button class="cell material-icons" gotag="name:UndoButton click:OnUndoButton" disabled>undo</button>
		<div class="cell-elastic footer-status"></div>
		<button class="cell material-icons" gotag="name:DoneButton click:OnDoneButton" disabled>done</button>
	</div>
</div>
`

type InitRemotePage struct {
	ctl *Controller

	root dom.Element

	BackButton, CloseButton, DoneButton, UndoButton *dom.HTMLButtonElement

	GitHost, GitUser, GitPass, GitPath *dom.HTMLInputElement

	GitServer *dom.HTMLSelectElement

	ToggleGitPassButton *dom.HTMLButtonElement
}

func NewInitRemotePage(ctl *Controller) (*InitRemotePage, error) {
	p := &InitRemotePage{
		ctl: ctl,
	}
	root, err := BindHTML(initRemoteTemplate, "gotag", p)
	if err != nil {
		return nil, xerrors.Errorf("could not parse init remote page template: %w", err)
	}
	p.root = root
	return p, nil
}

func (p *InitRemotePage) RootDiv() dom.Element {
	return p.root
}

func (p *InitRemotePage) RefreshDisplay() error {
	return p.RefreshButtons(nil)
}

func (p *InitRemotePage) OnBackButton(dom.Event) (status error) {
	if _, err := ShowSettingsPage(p.ctl); err != nil {
		return p.ctl.setError(err)
	}
	return nil
}

func (p *InitRemotePage) OnCloseButton(dom.Event) (status error) {
	return p.ctl.ClosePage(p)
}

func (p *InitRemotePage) OnDoneButton(dom.Event) (status error) {
	protocol := "git"
	if p := p.GitServer.Value(); p == "ssh" || p == "github-ssh" {
		protocol = "ssh"
	} else if p == "https" || p == "github-https" {
		protocol = "https"
	}
	req := &msg.AddRemoteRequest{
		Protocol:   protocol,
		Username:   p.GitUser.Value(),
		Password:   p.GitPass.Value(),
		Hostname:   p.GitHost.Value(),
		Path:       p.GitPath.Value(),
		SyncRemote: true,
	}
	resp, err := p.ctl.AddRemote(req)
	if err != nil {
		return p.ctl.setError(err)
	}
	page, err := NewRemotePage(p.ctl, resp.SyncRemote)
	if err != nil {
		return p.ctl.setError(err)
	}
	return p.ctl.ShowPage(page)
}

func (p *InitRemotePage) OnUndoButton(dom.Event) (status error) {
	p.GitServer.SetValue("ssh")
	p.GitHost.SetValue("")
	p.GitUser.SetValue("")
	p.GitPass.SetValue("")
	p.GitPath.SetValue("")
	p.UndoButton.SetDisabled(true)
	p.DoneButton.SetDisabled(true)
	return nil
}

func (p *InitRemotePage) OnToggleGitPass(dom.Event) (status error) {
	if strings.EqualFold(p.GitPass.Type(), "text") {
		p.GitPass.SetType("password")
		p.ToggleGitPassButton.SetTextContent(ShowSecretIcon)
	} else {
		p.GitPass.SetType("text")
		p.ToggleGitPassButton.SetTextContent(HideSecretIcon)
	}
	return nil
}

func (p *InitRemotePage) RefreshButtons(dom.Event) (status error) {
	p.DoneButton.SetDisabled(!p.isReady())
	p.UndoButton.SetDisabled(p.isEmpty())
	return nil
}

func (p *InitRemotePage) isReady() bool {
	emptyOK := false
	if v := p.GitServer.Value(); v == "ssh" || strings.HasSuffix(v, "-ssh") {
		emptyOK = true
	}
	if p.GitHost.Value() == "" || p.GitUser.Value() == "" || p.GitPath.Value() == "" {
		return false
	}
	if !emptyOK && p.GitPass.Value() == "" {
		return false
	}
	return true
}

func (p *InitRemotePage) isEmpty() bool {
	return p.GitHost.Value() == "" && p.GitUser.Value() == "" && p.GitPath.Value() == "" &&
		p.GitPass.Value() == "" && p.GitServer.Value() == "ssh"
}
