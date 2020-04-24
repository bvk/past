// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"golang.org/x/xerrors"
	"honnef.co/go/js/dom/v2"

	"github.com/bvk/past/msg"
)

const settingsTemplate = `
<div>
	<div class="row header">
		<button class="column material-icons" gotag="name:BackButton click:OnBackButton">navigate_before</button>
		<span class="column-elastic header-title">Settings</span>
		<button class="column material-icons" gotag="name:CloseButton click:OnCloseButton">clear</button>
	</div>

	<hr/>

	<div class="content mw32em">
		<ul>
			<li>
				<div class="row">
					<span class="column material-icons" gotag="name:MessagingIcon">clear</span>
					<span class="column-lefty">Chrome Native Messaging</span>
				</div>
			</li>

			<li>
				<div class="row">
					<span class="column material-icons" gotag="name:ToolsIcon">remove</span>
					<span class="column-lefty">Git and GPG Commands</span>
				</div>
			</li>

			<li>
				<div class="row">
					<span class="column material-icons" gotag="name:KeysIcon">remove</span>
					<span class="column-lefty">GPG Keyring</span>
					<button class="column material-icons" disabled gotag="name:KeysButton click:OnKeysButton">create_new_folder</button>
				</div>
			</li>

			<li>
				<div class="row">
					<span class="column material-icons" gotag="name:RepoIcon">remove</span>
					<span class="column-lefty">Password Store</span>
					<button class="column material-icons" disabled gotag="name:RepoButton click:OnRepoButton">create_new_folder</button>
				</div>
			</li>

			<li>
				<div class="row">
					<span class="column material-icons" gotag="name:RemoteIcon">remove</span>
					<span class="column-lefty">Remote Repository</span>
					<button class="column material-icons" disabled gotag="name:RemoteButton click:OnRemoteButton">create_new_folder</button>
				</div>
			</li>
		</ul>
	</div>

	<div class="row footer">
		<button class="column material-icons" gotag="name:CheckButton click:OnCheckButton">refresh</button>
		<div class="column-elastic footer-status"></div>
		<button class="column material-icons" gotag="name:CheckButton2 click:OnCheckButton">refresh</button>
	</div>
</div>
`

type SettingsPage struct {
	ctl *Controller

	args *msg.BrowserResponse

	root dom.Element

	BackButton, CloseButton *dom.HTMLButtonElement

	RepoButton, KeysButton, RemoteButton *dom.HTMLButtonElement

	CheckButton, CheckButton2 *dom.HTMLButtonElement

	MessagingIcon, ToolsIcon, KeysIcon, RepoIcon, RemoteIcon *dom.HTMLSpanElement
}

func NewSettingsPage(ctl *Controller, resp *msg.BrowserResponse) (*SettingsPage, error) {
	p := new(SettingsPage)
	root, err := BindHTML(settingsTemplate, "gotag", p)
	if err != nil {
		return nil, xerrors.Errorf("could not parse settings page template: %w", err)
	}
	p.ctl, p.args, p.root = ctl, resp, root
	return p, nil
}

func ShowSettingsPage(ctl *Controller) (*SettingsPage, error) {
	resp, err := ctl.CheckStatus()
	if err != nil {
		return nil, err
	}
	page, err := NewSettingsPage(ctl, resp)
	if err != nil {
		return nil, err
	}
	if err := ctl.ShowPage(page); err != nil {
		return nil, err
	}
	return page, nil
}

func (p *SettingsPage) RootDiv() dom.Element {
	return p.root
}

func (p *SettingsPage) RefreshDisplay() error {
	defer p.refreshBackButton()

	if p.args == nil || len(p.args.Status) > 0 {
		p.MessagingIcon.SetTextContent(NotReadyIcon)
		p.ToolsIcon.SetTextContent(NotReadyIcon)
		p.KeysIcon.SetTextContent(NotReadyIcon)
		p.RepoIcon.SetTextContent(NotReadyIcon)
		p.RemoteIcon.SetTextContent(NotReadyIcon)
		p.KeysButton.SetDisabled(true)
		p.RepoButton.SetDisabled(true)
		p.RemoteButton.SetDisabled(true)
		if len(p.args.Status) == 0 {
			p.ctl.setBadStatus("Not Ready")
		} else {
			p.ctl.setBadStatus(p.args.Status)
		}
		return nil
	}

	messagingReady := false
	if p.args.CheckStatus != nil {
		p.MessagingIcon.SetTextContent(ReadyIcon)
		messagingReady = true
	}

	toolsReady := false
	if messagingReady && len(p.args.CheckStatus.GitPath) > 0 && len(p.args.CheckStatus.GPGPath) > 0 {
		p.ToolsIcon.SetTextContent(ReadyIcon)
		toolsReady = true
	}

	keysReady := false
	if toolsReady && len(p.args.CheckStatus.LocalKeys) > 0 {
		p.KeysIcon.SetTextContent(ReadyIcon)
		p.KeysButton.SetTextContent(NavigateNextIcon)
		keysReady = true
	} else {
		p.KeysIcon.SetTextContent(NotReadyIcon)
		p.KeysButton.SetTextContent(CreateNewIcon)
	}
	p.KeysButton.SetDisabled(!toolsReady)

	repoReady := false
	if keysReady && len(p.args.CheckStatus.PasswordStoreKeys) > 0 {
		p.RepoIcon.SetTextContent(ReadyIcon)
		p.RepoButton.SetTextContent(NavigateNextIcon)
		repoReady = true
	} else {
		p.RepoIcon.SetTextContent(NotReadyIcon)
		p.RepoButton.SetTextContent(CreateNewIcon)
	}
	p.RepoButton.SetDisabled(!keysReady)

	remoteReady := false
	if repoReady && len(p.args.CheckStatus.Remote) != 0 {
		p.RemoteIcon.SetTextContent(ReadyIcon)
		p.RemoteButton.SetTextContent(NavigateNextIcon)
		remoteReady = true
	} else {
		p.RemoteIcon.SetTextContent(NotReadyIcon)
		p.RemoteButton.SetTextContent(CreateNewIcon)
	}
	p.RemoteButton.SetDisabled(!repoReady)
	_ = remoteReady

	if messagingReady && toolsReady && keysReady && repoReady {
		p.ctl.setGoodStatus("Ready")
	}

	return nil
}

func (p *SettingsPage) OnBackButton(dom.Event) (status error) {
	if _, err := ShowSearchPage(p.ctl); err != nil {
		return p.ctl.setError(err)
	}
	return nil
}

func (p *SettingsPage) OnCloseButton(dom.Event) (status error) {
	return p.ctl.ClosePage(p)
}

func (p *SettingsPage) OnRepoButton(dom.Event) (status error) {
	if !p.hasPasswordStore() {
		page, err := NewInitPastPage(p.ctl, p.args.CheckStatus)
		if err != nil {
			return p.ctl.setError(err)
		}
		if err := p.ctl.ShowPage(page); err != nil {
			return err
		}
		return nil
	}

	resp, err := p.ctl.PasswordStoreStatus()
	if err != nil {
		return p.ctl.setError(err)
	}
	page, err := NewPastPage(p.ctl, resp)
	if err != nil {
		return p.ctl.setError(err)
	}
	if err := p.ctl.ShowPage(page); err != nil {
		return err
	}
	return nil
}

func (p *SettingsPage) OnKeysButton(dom.Event) (status error) {
	if !p.hasKeyring() {
		page, err := NewAddKeyPage(p.ctl)
		if err != nil {
			return p.ctl.setError(err)
		}
		if err := p.ctl.ShowPage(page); err != nil {
			return err
		}
		return nil
	}
	page, err := NewKeyringPage(p.ctl, p.args.CheckStatus)
	if err != nil {
		return p.ctl.setError(err)
	}
	if err := p.ctl.ShowPage(page); err != nil {
		return err
	}
	return nil
}

func (p *SettingsPage) OnRemoteButton(dom.Event) (status error) {
	if !p.hasRemote() {
		page, err := NewInitRemotePage(p.ctl)
		if err != nil {
			return p.ctl.setError(err)
		}
		if err := p.ctl.ShowPage(page); err != nil {
			return err
		}
		return nil
	}
	resp, err := p.ctl.RemoteStatus()
	if err != nil {
		return p.ctl.setError(err)
	}
	page, err := NewRemotePage(p.ctl, resp)
	if err != nil {
		return p.ctl.setError(err)
	}
	if err := p.ctl.ShowPage(page); err != nil {
		return err
	}
	return nil
}

func (p *SettingsPage) OnCheckButton(dom.Event) (status error) {
	check, err := p.ctl.CheckStatus()
	if err != nil {
		return p.ctl.setError(err)
	}
	p.args = check
	if err := p.RefreshDisplay(); err != nil {
		return p.ctl.setError(err)
	}
	return nil
}

func (p *SettingsPage) refreshBackButton() {
	if p.isReady() {
		p.BackButton.SetDisabled(false)
	} else {
		p.BackButton.SetDisabled(true)
	}
}

func (p *SettingsPage) isReady() bool {
	return p.args != nil && len(p.args.Status) == 0 && p.args.CheckStatus != nil &&
		len(p.args.CheckStatus.GitPath) != 0 && len(p.args.CheckStatus.GPGPath) != 0 &&
		len(p.args.CheckStatus.LocalKeys) != 0 &&
		len(p.args.CheckStatus.PasswordStoreKeys) != 0
}

func (p *SettingsPage) hasTools() bool {
	return p.args != nil && len(p.args.Status) == 0 && p.args.CheckStatus != nil &&
		len(p.args.CheckStatus.GitPath) != 0 && len(p.args.CheckStatus.GPGPath) != 0
}

func (p *SettingsPage) hasKeyring() bool {
	return p.args != nil && len(p.args.Status) == 0 && p.args.CheckStatus != nil &&
		len(p.args.CheckStatus.LocalKeys) != 0
}

func (p *SettingsPage) hasPasswordStore() bool {
	return p.args != nil && len(p.args.Status) == 0 && p.args.CheckStatus != nil &&
		len(p.args.CheckStatus.PasswordStoreKeys) != 0
}

func (p *SettingsPage) hasRemote() bool {
	return p.args != nil && len(p.args.Status) == 0 && p.args.CheckStatus != nil &&
		len(p.args.CheckStatus.Remote) != 0
}
