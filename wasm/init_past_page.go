// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"strings"

	"golang.org/x/xerrors"
	"honnef.co/go/js/dom/v2"

	"github.com/bvk/past"
	"github.com/bvk/past/msg"
)

const initPastTemplate = `
<div>
	<div class="row header">
		<button class="cell material-icons" gotag="name:BackButton click:OnBackButton">navigate_before</button>
		<span class="cell-elastic header-title">New Password Store</span>
		<button class="cell material-icons" gotag="name:CloseButton click:OnCloseButton">clear</button>
	</div>

	<div class="content mw32em">
		<div class="row tab-bar">
			<button class="cell-elastic" gotag="name:ImportButton click:OnImportButton">
				<span class="button-text material-icons">cloud_download</span>
				<span class="button-text">Import</span>
			</button>
			<button class="cell-elastic" gotag="name:CreateButton click:OnCreateButton">
				<span class="button-text material-icons">create_new_folder</span>
				<span class="button-text">Create</span>
			</button>
		</div>

		<hr/>

		<div gotag="name:CreateTab">
			<div class="localkeys-section">
				<div class="row">
					<span class="cell-elastic bold">Local Keys</span>
				</div>
				<ul>
					<li gotag="name:LocalKeyTemplate" style="display:none">
						<div class="row">
							<span class="cell material-icons">360</span>
							<span class="cell-lefty" localkey="name:KeyID click:OnKeyID"></span>
							<button class="cell material-icons" localkey="name:KeyCheckbox click:OnKeyCheckbox">check_box_outline_blank</button>
						</div>
					</li>
				</ul>
			</div>

			<div class="remotekeys-section">
				<div class="row">
					<span class="cell-elastic bold">Remote Keys</span>
				</div>
				<ul>
					<li gotag="name:RemoteKeyTemplate" style="display:none">
						<div class="row">
							<span class="cell material-icons">360</span>
							<span class="cell-lefty" remotekey="name:KeyID click:OnKeyID"></span>
							<button class="cell material-icons" remotekey="name:KeyCheckbox click:OnKeyCheckbox">check_box_outline_blank</button>
						</div>
					</li>
				</ul>
			</div>
			<div class="expiredkeys-section">
				<div class="row">
					<span class="cell-elastic bold">Expired Keys</span>
				</div>
				<ul>
					<li gotag="name:ExpiredKeyTemplate" style="display:none">
						<div class="row">
							<span class="cell material-icons">360</span>
							<span class="cell-lefty" expiredkey="name:KeyID click:OnKeyID"></span>
							<button class="cell material-icons" expiredkey="name:KeyCheckbox click:OnKeyCheckbox" disabled>check_box_outline_blank</button>
						</div>
					</li>
				</ul>
			</div>
		</div>

		<div gotag="name:ImportTab">
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

			<!-- TODO: Add a checkbox for ssh-key fingerprint acceptance -->

		</div>
	</div>

	<div class="row footer">
		<button class="cell material-icons" gotag="name:UndoButton click:OnUndoButton" disabled>undo</button>
		<div class="cell-elastic footer-status"></div>
		<button class="cell material-icons" gotag="name:DoneButton click:OnDoneButton" disabled>done</button>
	</div>
</div>
`

type InitPastPage struct {
	ctl *Controller

	args *msg.CheckStatusResponse

	currentTab string

	selectedKeys map[string]struct{}

	root dom.Element

	BackButton, CloseButton, UndoButton, DoneButton *dom.HTMLButtonElement

	CreateTab, ImportTab *dom.HTMLDivElement

	CreateButton, ImportButton *dom.HTMLButtonElement

	GitServer *dom.HTMLSelectElement

	GitHost, GitUser, GitPass, GitPath *dom.HTMLInputElement

	ToggleGitPassButton *dom.HTMLButtonElement

	LocalKeyTemplate, RemoteKeyTemplate, ExpiredKeyTemplate *dom.HTMLLIElement
}

func NewInitPastPage(ctl *Controller, args *msg.CheckStatusResponse) (*InitPastPage, error) {
	p := &InitPastPage{
		ctl:          ctl,
		args:         args,
		currentTab:   "create-tab",
		selectedKeys: make(map[string]struct{}),
	}
	root, err := BindHTML(initPastTemplate, "gotag", p)
	if err != nil {
		return nil, xerrors.Errorf("could not parse init password store page template: %w", err)
	}
	p.root = root
	return p, nil
	return p, nil
}

func (p *InitPastPage) RootDiv() dom.Element {
	return p.root
}

func (p *InitPastPage) RefreshDisplay() error {
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

		k := &pastKeyListItem{page: p, key: key}
		if err := BindTarget(clone, "localkey", k); err != nil {
			return p.ctl.setError(err)
		}
		k.KeyID.SetTextContent(key.UserName)

		clone.Style().RemoveProperty("display")
		p.LocalKeyTemplate.ParentNode().InsertBefore(clone, p.LocalKeyTemplate.NextSibling())
	}
	for _, key := range p.args.RemoteKeys {
		clone := mustCloneHTMLElement(p.RemoteKeyTemplate, true)

		k := &pastKeyListItem{page: p, key: key}
		if err := BindTarget(clone, "remotekey", k); err != nil {
			return p.ctl.setError(err)
		}
		k.KeyID.SetTextContent(key.UserName)
		if !key.IsTrusted {
			k.KeyCheckbox.SetDisabled(true)
		}

		clone.Style().RemoveProperty("display")
		p.RemoteKeyTemplate.ParentNode().InsertBefore(clone, p.RemoteKeyTemplate.NextSibling())
	}
	for _, key := range p.args.ExpiredKeys {
		clone := mustCloneHTMLElement(p.ExpiredKeyTemplate, true)

		k := pastKeyListItem{page: p, key: key}
		if err := BindTarget(clone, "expiredkey", k); err != nil {
			return p.ctl.setError(err)
		}
		k.KeyID.SetTextContent(key.UserName)

		clone.Style().RemoveProperty("display")
		p.ExpiredKeyTemplate.ParentNode().InsertBefore(clone, p.ExpiredKeyTemplate.NextSibling())
	}

	// Display the current tab and hide other tabs.
	type Tab struct {
		name   string
		button *dom.HTMLButtonElement
		tab    *dom.HTMLDivElement
	}
	var tabs = []Tab{
		{"create-tab", p.CreateButton, p.CreateTab},
		{"import-tab", p.ImportButton, p.ImportTab},
	}
	for _, tab := range tabs {
		if tab.name == p.currentTab {
			tab.button.Style().SetProperty("background", "gray", "")
			tab.tab.Style().RemoveProperty("display")
		} else {
			tab.button.Style().SetProperty("background", "transparent", "")
			tab.tab.Style().SetProperty("display", "none", "")
		}
	}
	return p.RefreshButtons(nil)
}

func (p *InitPastPage) OnCreateButton(dom.Event) (status error) {
	p.currentTab = "create-tab"
	return p.RefreshDisplay()
}

func (p *InitPastPage) OnImportButton(dom.Event) (status error) {
	p.currentTab = "import-tab"
	return p.RefreshDisplay()
}

func (p *InitPastPage) OnToggleGitPass(dom.Event) (status error) {
	if strings.EqualFold(p.GitPass.Type(), "text") {
		p.GitPass.SetType("password")
		p.ToggleGitPassButton.SetTextContent(ShowSecretIcon)
	} else {
		p.GitPass.SetType("text")
		p.ToggleGitPassButton.SetTextContent(HideSecretIcon)
	}
	return nil
}

func (p *InitPastPage) OnBackButton(dom.Event) (status error) {
	if _, err := ShowSettingsPage(p.ctl); err != nil {
		return p.ctl.setError(err)
	}
	return nil
}

func (p *InitPastPage) OnCloseButton(dom.Event) (status error) {
	return p.ctl.ClosePage(p)
}

func (p *InitPastPage) OnDoneButton(dom.Event) (status error) {
	var checkStatus *msg.CheckStatusResponse
	if p.currentTab == "import-tab" {
		protocol := "git"
		if p := p.GitServer.Value(); p == "ssh" || p == "github-ssh" {
			protocol = "ssh"
		} else if p == "https" || p == "github-https" {
			protocol = "https"
		}
		req := &msg.ImportRepoRequest{
			Protocol:    protocol,
			Username:    p.GitUser.Value(),
			Password:    p.GitPass.Value(),
			Hostname:    p.GitHost.Value(),
			Path:        p.GitPath.Value(),
			CheckStatus: true,
		}
		resp, err := p.ctl.ImportRepo(req)
		if err != nil {
			return p.ctl.setError(err)
		}
		checkStatus = resp.CheckStatus
	} else { // create-tab
		req := &msg.CreateRepoRequest{
			CheckStatus: true,
		}
		for fp := range p.selectedKeys {
			req.Fingerprints = append(req.Fingerprints, fp)
		}
		resp, err := p.ctl.CreateRepo(req)
		if err != nil {
			return p.ctl.setError(err)
		}
		checkStatus = resp.CheckStatus
	}

	params := &msg.Response{
		CheckStatus: checkStatus,
	}
	page, err := NewSettingsPage(p.ctl, params)
	if err != nil {
		return p.ctl.setError(err)
	}
	return p.ctl.ShowPage(page)
}

func (p *InitPastPage) OnUndoButton(dom.Event) (status error) {
	if p.currentTab == "import-tab" {
		p.GitServer.SetValue("ssh")
		p.GitHost.SetValue("")
		p.GitUser.SetValue("")
		p.GitPass.SetValue("")
		p.GitPath.SetValue("")
	} else { // create-tab
		checks := getButtonsByClassName(p.root, "localkey-checkbox")
		for _, check := range checks {
			check.SetTextContent(UncheckedIcon)
		}
		p.selectedKeys = make(map[string]struct{})
	}
	p.UndoButton.SetDisabled(true)
	p.DoneButton.SetDisabled(true)
	return nil
}

func (p *InitPastPage) RefreshButtons(dom.Event) (status error) {
	if p.currentTab == "import-tab" {
		p.DoneButton.SetDisabled(!p.isImportReady())
		p.UndoButton.SetDisabled(p.isImportEmpty())
	} else { // create-tab
		p.DoneButton.SetDisabled(!p.isCreateReady())
		p.UndoButton.SetDisabled(p.isCreateEmpty())
	}
	return nil
}

func (p *InitPastPage) isImportReady() bool {
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

func (p *InitPastPage) isImportEmpty() bool {
	return p.GitHost.Value() == "" && p.GitUser.Value() == "" && p.GitPath.Value() == "" &&
		p.GitPass.Value() == "" && p.GitServer.Value() == "ssh"
}

func (p *InitPastPage) isCreateReady() bool {
	return len(p.selectedKeys) > 0
}

func (p *InitPastPage) isCreateEmpty() bool {
	return len(p.selectedKeys) == 0
}

type pastKeyListItem struct {
	page *InitPastPage

	key *past.PublicKeyData

	KeyID *dom.HTMLSpanElement

	KeyCheckbox *dom.HTMLButtonElement

	ViewKey *dom.HTMLButtonElement
}

func (k *pastKeyListItem) OnKeyID(dom.Event) (status error) {
	rotateKeyID(k.KeyID, k.key)
	return nil
}

func (k *pastKeyListItem) OnViewKey(dom.Event) (status error) {
	page, err := NewKeyPage(k.page.ctl, k.key)
	if err != nil {
		return k.page.ctl.setError(err)
	}
	return k.page.ctl.ShowPage(page)
}

func (k *pastKeyListItem) OnKeyCheckbox(dom.Event) (status error) {
	if _, ok := k.page.selectedKeys[k.key.KeyFingerprint]; ok {
		k.KeyCheckbox.SetTextContent(UncheckedIcon)
		delete(k.page.selectedKeys, k.key.KeyFingerprint)
	} else {
		k.KeyCheckbox.SetTextContent(CheckedIcon)
		k.page.selectedKeys[k.key.KeyFingerprint] = struct{}{}
	}
	return nil
}
