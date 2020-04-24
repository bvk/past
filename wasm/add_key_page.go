// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"strings"

	"golang.org/x/xerrors"
	"honnef.co/go/js/dom/v2"

	"github.com/bvk/past/msg"
)

const addKeyTemplate = `
<div>
	<div class="row header">
		<button class="column material-icons" gotag="name:BackButton click:OnBackButton">navigate_before</button>
		<span class="column-elastic header-title">Add Encryption Keys</span>
		<button class="column material-icons" gotag="name:CloseButton click:OnCloseButton">clear</button>
	</div>

	<div class="content mw32em">
		<div class="row tab-bar">
			<button class="column-elastic" gotag="name:CreateTabButton click:OnCreateTabButton">
				<span class="button-text material-icons">create_new_folder</span>
				<span class="button-text">Create</span>
			</button>

			<button class="column-elastic" gotag="name:ImportTabButton click:OnImportTabButton">
				<span class="button-text material-icons">cloud_download</span>
				<span class="button-text">Import</span>
			</button>
		</div>

		<hr/>

		<div gotag="name:CreateTab">
			<div class="row">
				<span class="column w6em">User Name</span>
				<input class="column-elastic" gotag="name:UserName input:RefreshButtons"></input>
			</div>

			<div class="row">
				<span class="column w6em">User Email</span>
				<input class="column-elastic" gotag="name:UserEmail input:RefreshButtons"></input>
			</div>

			<div class="row">
				<span class="column w6em">Key Options</span>
				<select class="column-elastic" gotag="name:KeyLength change:RefreshButtons">
					<option value="4096">4096 Bits</option>
					<option value="2048">2048 Bits</option>
					<option value="1024">1024 Bits</option>
				</select>
				<select class="column-elastic" gotag="name:KeyYears change:RefreshButtons">
					<option value="0">Doesn't Expire</option>
					<option value="1">1 Year</option>
					<option value="2">2 Years</option>
					<option value="3">3 Years</option>
					<option value="4">4 Years</option>
					<option value="5">5 Years</option>
					<option value="10">10 Years</option>
				</select>
			</div>

			<div class="row">
				<span class="column-elastic">Passphrase</span>
			</div>

			<div class="row">
				<input class="column-elastic w4em" type="password" gotag="name:Passphrase input:RefreshButtons"></input>
				<input class="column-elastic w4em" type="password" gotag="name:RepeatPassphrase input:RefreshButtons"></input>
				<button class="column material-icons" gotag="name:ToggleButton click:OnToggleButton">visibility_off</button>
			</div>
		</div>

		<div class="row" gotag="name:ImportTab">
			<textarea class="column-elastic h25em" gotag="name:KeyData input:RefreshButtons"></textarea>
		</div>
	</div>

	<div class="row footer">
		<button class="column material-icons" disabled gotag="name:UndoButton click:OnUndoButton">undo</button>
		<div class="column-elastic footer-status"></div>
		<button class="column material-icons" disabled gotag="name:DoneButton click:OnDoneButton">done</button>
	</div>
</div>
`

type AddKeyPage struct {
	ctl *Controller

	root dom.Element

	currentTab string

	BackButton, CloseButton *dom.HTMLButtonElement
	DoneButton, UndoButton  *dom.HTMLButtonElement

	CreateTabButton, ImportTabButton *dom.HTMLButtonElement

	CreateTab, ImportTab *dom.HTMLDivElement

	ToggleButton *dom.HTMLButtonElement

	UserName, UserEmail, Passphrase, RepeatPassphrase *dom.HTMLInputElement

	KeyLength, KeyYears *dom.HTMLSelectElement

	KeyData *dom.HTMLTextAreaElement
}

func NewAddKeyPage(ctl *Controller) (*AddKeyPage, error) {
	p := &AddKeyPage{
		ctl:        ctl,
		currentTab: "create-tab",
	}
	root, err := BindHTML(addKeyTemplate, "gotag", p)
	if err != nil {
		return nil, xerrors.Errorf("could not parse add key page template: %w", err)
	}
	p.root = root
	return p, nil
}

func (p *AddKeyPage) RefreshDisplay() error {
	type Tab struct {
		name    string
		button  *dom.HTMLButtonElement
		content dom.HTMLElement
	}
	var tabs = []Tab{
		{"create-tab", p.CreateTabButton, p.CreateTab},
		{"import-tab", p.ImportTabButton, p.ImportTab},
	}
	for _, tab := range tabs {
		if tab.name == p.currentTab {
			tab.button.Style().SetProperty("background", "gray", "")
			tab.content.Style().RemoveProperty("display")
		} else {
			tab.button.Style().SetProperty("background", "transparent", "")
			tab.content.Style().SetProperty("display", "none", "")
		}
	}
	return p.RefreshButtons(nil)
}

func (p *AddKeyPage) RootDiv() dom.Element {
	return p.root
}

func (p *AddKeyPage) OnCreateTabButton(dom.Event) (status error) {
	p.currentTab = "create-tab"
	return p.RefreshDisplay()
}

func (p *AddKeyPage) OnImportTabButton(dom.Event) (status error) {
	p.currentTab = "import-tab"
	return p.RefreshDisplay()
}

func (p *AddKeyPage) OnCloseButton(dom.Event) (status error) {
	return p.ctl.ClosePage(p)
}

func (p *AddKeyPage) RefreshButtons(dom.Event) (status error) {
	if p.currentTab == "import-tab" {
		if v := p.KeyData.Value(); len(v) == 0 {
			p.DoneButton.SetDisabled(true)
			p.UndoButton.SetDisabled(true)
		} else {
			p.DoneButton.SetDisabled(false)
			p.UndoButton.SetDisabled(false)
		}
		return nil
	}
	// Create tab.
	if p.isCreateReady() {
		p.DoneButton.SetDisabled(false)
	} else {
		p.DoneButton.SetDisabled(true)
	}
	if p.isCreateEmpty() {
		p.UndoButton.SetDisabled(true)
	} else {
		p.UndoButton.SetDisabled(false)
	}
	return nil
}

func (p *AddKeyPage) isCreateReady() bool {
	return p.UserName.Value() != "" && p.UserEmail.Value() != "" && p.Passphrase.Value() != "" &&
		p.Passphrase.Value() == p.RepeatPassphrase.Value()
}

func (p *AddKeyPage) isCreateEmpty() bool {
	return p.UserName.Value() == "" && p.UserEmail.Value() == "" && p.Passphrase.Value() == "" &&
		p.RepeatPassphrase.Value() == "" && p.KeyLength.Value() == "4096" && p.KeyYears.Value() == "0"
}

func (p *AddKeyPage) OnBackButton(dom.Event) (status error) {
	resp, err := p.ctl.CheckStatus()
	if err != nil {
		return p.ctl.setError(err)
	}
	if len(resp.CheckStatus.LocalKeys) == 0 {
		page, err := NewSettingsPage(p.ctl, resp)
		if err != nil {
			return p.ctl.setError(err)
		}
		return p.ctl.ShowPage(page)
	}
	page, err := NewKeyringPage(p.ctl, resp.CheckStatus)
	if err != nil {
		return p.ctl.setError(err)
	}
	return p.ctl.ShowPage(page)
}

func (p *AddKeyPage) OnDoneButton(dom.Event) (status error) {
	var newStatus *msg.CheckStatusResponse
	if p.currentTab == "import-tab" {
		req := &msg.ImportKeyRequest{
			CheckStatus: true,
			Key:         p.KeyData.Value(),
		}
		resp, err := p.ctl.ImportKey(req)
		if err != nil {
			return p.ctl.setError(err)
		}
		newStatus = resp.CheckStatus
	} else { // create tab
		req := &msg.CreateKeyRequest{
			Name:        p.UserName.Value(),
			Email:       p.UserEmail.Value(),
			Passphrase:  p.Passphrase.Value(),
			KeyLength:   mustAtoi(p.KeyLength.Value()),
			KeyYears:    mustAtoi(p.KeyYears.Value()),
			CheckStatus: true,
		}

		bs := StartBusy(p.root)
		resp, err := p.ctl.CreateKey(req)
		StopBusy(bs)

		if err != nil {
			return p.ctl.setError(err)
		}
		newStatus = resp.CheckStatus
	}
	page, err := NewKeyringPage(p.ctl, newStatus)
	if err != nil {
		return p.ctl.setError(err)
	}
	if err := p.ctl.ShowPage(page); err != nil {
		return err
	}
	return nil
}

func (p *AddKeyPage) OnUndoButton(dom.Event) (status error) {
	if p.currentTab == "import-tab" {
		p.KeyData.SetValue("")
	} else {
		p.UserName.SetValue("")
		p.UserEmail.SetValue("")
		p.Passphrase.SetValue("")
		p.RepeatPassphrase.SetValue("")
		p.KeyLength.SetValue("4096")
		p.KeyYears.SetValue("0")
	}
	p.UndoButton.SetDisabled(true)
	p.DoneButton.SetDisabled(true)
	return nil
}

func (p *AddKeyPage) OnToggleButton(dom.Event) (status error) {
	if strings.EqualFold(p.Passphrase.Type(), "text") {
		p.Passphrase.SetType("password")
		p.RepeatPassphrase.SetType("password")
		p.ToggleButton.SetTextContent(HideSecretIcon)
	} else {
		p.Passphrase.SetType("text")
		p.RepeatPassphrase.SetType("text")
		p.ToggleButton.SetTextContent(ShowSecretIcon)
	}
	return nil
}
