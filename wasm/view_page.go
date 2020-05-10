// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"time"

	"golang.org/x/xerrors"
	"honnef.co/go/js/dom/v2"

	"github.com/bvk/past/msg"
)

const viewTemplate = `
<div>
	<div class="row header">
		<button class="cell material-icons" gotag="name:BackButton click:OnBackButton">navigate_before</button>
		<span class="cell-elastic header-title" gotag="name:Filename">Password File Name</span>
		<button class="cell material-icons" gotag="name:CloseButton click:OnCloseButton">clear</button>
	</div>

	<hr/>

	<div class="content mw32em">
		<div class="row">
			<span class="cell w6em">Username</span>
			<input class="cell-elastic" gotag="name:Username" readonly></input>
			<button class="cell material-icons" gotag="name:CopyUsername click:OnCopyUsername">content_copy</button>
		</div>

		<div class="row">
			<span class="cell w6em">Password</span>
			<input class="cell-elastic" gotag="name:Password" type="password" readonly></input>
			<button class="cell material-icons" gotag="name:TogglePassword click:OnTogglePassword">visibility_off</button>
			<button class="cell material-icons" gotag="name:CopyPassword click:OnCopyPassword">content_copy</button>
		</div>

		<div>Other data</div>

		<div class="row">
			<textarea class="cell-elastic h4em" gotag="name:UserData" readonly></textarea>
		</div>
	</div>

	<div class="row footer">
		<button class="cell material-icons" gotag="name:DeleteButton click:OnDeleteButton">delete</button>
		<div class="cell-elastic footer-status"></div>
		<button class="cell material-icons" gotag="name:EditButton click:OnEditButton">edit</button>
	</div>
</div>
`

type ViewPage struct {
	ctl *Controller

	data *msg.ViewFileResponse

	root dom.Element

	BackButton, CloseButton, EditButton, DeleteButton *dom.HTMLButtonElement

	CopyUsername, CopyPassword, TogglePassword *dom.HTMLButtonElement

	Filename *dom.HTMLSpanElement

	Username, Password *dom.HTMLInputElement

	UserData *dom.HTMLTextAreaElement
}

func NewViewPage(ctl *Controller, data *msg.ViewFileResponse) (*ViewPage, error) {
	p := &ViewPage{
		ctl:  ctl,
		data: data,
	}
	root, err := BindHTML(viewTemplate, "gotag", p)
	if err != nil {
		return nil, xerrors.Errorf("could not parse view page template: %w", err)
	}
	p.root = root
	return p, nil
}

func ShowViewPage(ctl *Controller, filename string) (*ViewPage, error) {
	req := msg.ViewFileRequest{
		Filename: filename,
	}
	resp, err := ctl.ViewFile(&req)
	if err != nil {
		return nil, ctl.setError(err)
	}
	page, err := NewViewPage(ctl, resp)
	if err != nil {
		return nil, ctl.setError(err)
	}
	if err := ctl.ShowPage(page); err != nil {
		return nil, err
	}
	return nil, err
}

func (p *ViewPage) RootDiv() dom.Element {
	return p.root
}

func (p *ViewPage) RefreshDisplay() error {
	p.Filename.SetTextContent(p.data.Filename)
	p.Username.SetValue(p.data.Username)
	p.Password.SetValue(p.data.Password)
	p.UserData.SetValue(p.data.Data)
	return nil
}

func (p *ViewPage) OnBackButton(dom.Event) (status error) {
	if _, err := ShowSearchPage(p.ctl); err != nil {
		return p.ctl.setError(err)
	}
	return nil
}

func (p *ViewPage) OnCloseButton(dom.Event) (status error) {
	return p.ctl.ClosePage(p)
}

func (p *ViewPage) OnEditButton(dom.Event) (status error) {
	page, err := NewEditPage(p.ctl, p.data)
	if err != nil {
		return p.ctl.setError(err)
	}
	if err := p.ctl.ShowPage(page); err != nil {
		return err
	}
	return nil
}

func (p *ViewPage) OnDeleteButton(dom.Event) (status error) {
	req := &msg.DeleteFileRequest{
		File: p.data.Filename,
	}
	if _, err := p.ctl.DeleteFile(req); err != nil {
		return p.ctl.setError(err)
	}
	return p.OnBackButton(nil)
}

func (p *ViewPage) OnTogglePassword(dom.Event) (status error) {
	if t := p.Password.Type(); t == "text" {
		p.Password.SetType("password")
		p.TogglePassword.SetTextContent(HideSecretIcon)
	} else {
		p.Password.SetType("text")
		p.TogglePassword.SetTextContent(ShowSecretIcon)
	}
	return nil
}

func (p *ViewPage) OnCopyUsername(dom.Event) (status error) {
	if err := p.ctl.Copy(p.data.Username); err != nil {
		return p.ctl.setError(err)
	}
	return nil
}

func (p *ViewPage) OnCopyPassword(dom.Event) (status error) {
	if err := p.ctl.CopyTimeout(p.data.Password, 10*time.Second); err != nil {
		return p.ctl.setError(err)
	}
	return nil
}
