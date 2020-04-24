// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/xerrors"
	"honnef.co/go/js/dom/v2"

	"github.com/bvk/past/msg"
)

const editTemplate = `
<div>
	<div class="row header">
		<button class="column material-icons" gotag="name:BackButton click:OnBackButton">navigate_before</button>
		<span class="column-elastic header-title" gotag="name:Title">New Password File</span>
		<button class="column material-icons" gotag="name:CloseButton click:OnCloseButton">clear</button>
	</div>

	<hr/>

	<div class="content mw32em">
		<div>
			<div class="row">
				<div class="column w6em">Website</div>
				<input class="column-lefty" gotag="name:Sitename input:RefreshButtons" type="text" placeholder="website name"></input>
			</div>

			<div class="row">
				<span class="column w6em">Username</span>
				<input class="column-lefty" gotag="name:Username input:RefreshButtons" type="text" placeholder="account name"></input>
			</div>

			<div class="row">
				<span class="column w6em">Password</span>
				<select class="column-lefty" gotag="name:PasswordType change:OnPasswordType">
					<option value="">User Typed Password</option>
					<option value="letters_numbers_symbols">Letters, Numbers, Symbols</option>
					<option value="letters_numbers">Letters, Numbers</option>
					<option value="numbers">Numbers</option>
					<option value="letters">Letters</option>
					<option value="base64std">Standard Base64</option>
				</select>
				<span class="column" gotag="name:PasswordSize wheel:OnPasswordSize" style="text-decoration: line-through">12 Runes</span>
			</div>

			<div class="row">
				<input class="column-elastic password-element" gotag="name:Password input:RefreshButtons" type="password" placeholder="password"></input>
				<input class="column-elastic password-element" gotag="name:RepeatPassword input:RefreshButtons" type="password" placeholder="retype password"></input>
				<button class="column material-icons" gotag="name:TogglePassword click:OnTogglePassword">visibility_off</button>
				<button class="column material-icons" gotag="name:CopyPassword click:OnCopyPassword" disabled>content_copy</button>
				<button class="column material-icons" gotag="name:GeneratePassword click:OnGeneratePassword" disabled>refresh</button>
			</div>

			<div>Other data</div>
			<div class="row">
				<textarea class="column-elastic h4em" gotag="name:UserData input:RefreshButtons"></textarea>
			</div>
		</div>
	</div>

	<div class="row footer">
		<button class="column material-icons" gotag="name:UndoButton click:OnUndoButton">undo</button>
		<div class="column-elastic footer-status"></div>
		<button class="column material-icons" gotag="name:DoneButton click:OnDoneButton" disabled>done</button>
	</div>
</div>
`

type EditPage struct {
	ctl *Controller

	data *msg.ViewFileResponse

	root dom.Element

	BackButton, CloseButton, UndoButton, DoneButton *dom.HTMLButtonElement

	Sitename, Username, Password, RepeatPassword *dom.HTMLInputElement

	UserData *dom.HTMLTextAreaElement

	PasswordType *dom.HTMLSelectElement

	Title, PasswordSize *dom.HTMLSpanElement

	CopyPassword, TogglePassword, GeneratePassword *dom.HTMLButtonElement
}

func NewEditPage(ctl *Controller, data *msg.ViewFileResponse) (*EditPage, error) {
	p := &EditPage{
		ctl:  ctl,
		data: data,
	}
	root, err := BindHTML(editTemplate, "gotag", p)
	if err != nil {
		return nil, xerrors.Errorf("could not parse edit page template: %w", err)
	}
	p.root = root
	return p, nil
}

func (p *EditPage) RootDiv() dom.Element {
	return p.root
}

func (p *EditPage) RefreshDisplay() error {
	p.Title.SetTextContent(p.getTitle())
	p.Sitename.SetValue(p.data.Sitename)
	p.Username.SetValue(p.data.Username)
	p.Password.SetValue(p.data.Password)
	p.UserData.SetValue(p.data.Data)
	p.RepeatPassword.SetValue(p.data.Password)

	return p.RefreshButtons(nil)
}

func (p *EditPage) OnBackButton(dom.Event) (status error) {
	if _, err := ShowSearchPage(p.ctl); err != nil {
		return p.ctl.setError(err)
	}
	return nil
}

func (p *EditPage) OnCloseButton(dom.Event) (status error) {
	return p.ctl.ClosePage(p)
}

func (p *EditPage) OnDoneButton(dom.Event) (status error) {
	if p.data.Filename == "" {
		req := &msg.AddFileRequest{
			Sitename: p.Sitename.Value(),
			Username: p.Username.Value(),
			Password: p.Password.Value(),
			Data:     p.UserData.Value(),
		}
		if _, err := p.ctl.AddFile(req); err != nil {
			return p.ctl.setError(err)
		}
		return p.OnBackButton(nil)
	}

	req := &msg.EditFileRequest{
		OrigFile: p.data.Filename,
		Sitename: p.Sitename.Value(),
		Username: p.Username.Value(),
		Password: p.Password.Value(),
		Data:     p.UserData.Value(),
	}
	if _, err := p.ctl.EditFile(req); err != nil {
		return p.ctl.setError(err)
	}
	return p.OnBackButton(nil)
}

func (p *EditPage) OnUndoButton(dom.Event) (status error) {
	// Reset all modifiable items to their initial state.
	p.Sitename.SetValue(p.data.Sitename)
	p.Username.SetValue(p.data.Username)
	p.Password.SetValue(p.data.Password)
	p.UserData.SetValue(p.data.Data)
	p.RepeatPassword.SetValue(p.data.Password)

	p.PasswordType.SetValue("")
	p.Password.SetDisabled(false)
	p.RepeatPassword.SetDisabled(false)
	p.GeneratePassword.SetDisabled(true)
	p.PasswordSize.SetTextContent("12 Runes")
	p.PasswordSize.Style().SetProperty("text-decoration", "line-through", "")

	p.UndoButton.SetDisabled(true)
	p.DoneButton.SetDisabled(true)
	return nil
}

func (p *EditPage) OnCopyPassword(dom.Event) (status error) {
	return p.ctl.CopyTimeout(p.Password.Value(), 10*time.Second)
}

func (p *EditPage) OnPasswordType(dom.Event) (status error) {
	if p.PasswordType.Value() == "" {
		// User typed password
		p.Password.SetDisabled(false)
		p.RepeatPassword.SetDisabled(false)
		p.Password.SetValue("")
		p.RepeatPassword.SetValue("")
		p.GeneratePassword.SetDisabled(true)
		p.PasswordSize.Style().SetProperty("text-decoration", "line-through", "")
	} else {
		p.Password.SetDisabled(true)
		p.RepeatPassword.SetDisabled(true)
		p.GeneratePassword.SetDisabled(false)
		p.PasswordSize.Style().RemoveProperty("text-decoration")
		if err := p.OnGeneratePassword(nil); err != nil {
			return err
		}
	}
	return p.RefreshButtons(nil)
}

func (p *EditPage) OnTogglePassword(dom.Event) (status error) {
	if t := p.Password.Type(); t == "text" {
		p.Password.SetType("password")
		p.RepeatPassword.SetType("password")
		p.TogglePassword.SetTextContent(HideSecretIcon)
	} else {
		p.Password.SetType("text")
		p.RepeatPassword.SetType("text")
		p.TogglePassword.SetTextContent(ShowSecretIcon)
	}
	return nil
}

func (p *EditPage) OnPasswordSize(e dom.Event) error {
	we, ok := e.(*dom.WheelEvent)
	if !ok {
		return xerrors.Errorf("event is not a WheelEvent: %w", os.ErrInvalid)
	}
	words := strings.Fields(p.PasswordSize.TextContent())
	value := mustAtoi(words[0])
	if value < 32 && we.DeltaY() > 0 {
		value += 1
	} else if value > 3 && we.DeltaY() < 0 {
		value -= 1
	}
	words[0] = fmt.Sprintf("%d", value)
	p.PasswordSize.SetTextContent(strings.Join(words, " "))
	return p.OnGeneratePassword(nil)
}

func (p *EditPage) OnGeneratePassword(dom.Event) (status error) {
	words := strings.Fields(p.PasswordSize.TextContent())
	length := mustAtoi(words[0])

	var (
		numbers   = "0123456789"
		letters   = "ABCDEFGHIJKLMNOPQRSTUVWXZYabcdefghijklmnopqrstuvwxzy"
		symbols   = `"!#$%&'()*+,\-./:;<=>?@[]^_{|}~` + "`"
		base64std = letters + numbers + "+/"
	)

	password := ""
	switch p.PasswordType.Value() {
	case "numbers":
		password = pwgen(numbers, length)
	case "letters":
		password = pwgen(letters, length)
	case "letters_numbers":
		password = pwgen(letters+numbers, length)
	case "letters_numbers_symbols":
		password = pwgen(letters+numbers+symbols, length)
	case "base64std":
		password = pwgen(base64std, length)
	}

	p.Password.SetValue(password)
	p.RepeatPassword.SetValue(password)
	return nil
}

func (p *EditPage) RefreshButtons(dom.Event) (status error) {
	if p1, p2 := p.Password.Value(), p.RepeatPassword.Value(); p1 != p2 {
		p.CopyPassword.SetDisabled(true)
	} else {
		p.CopyPassword.SetDisabled(false)
	}
	if p.isEmpty() {
		p.UndoButton.SetDisabled(true)
	} else {
		p.UndoButton.SetDisabled(false)
	}
	if p.isReady() {
		p.DoneButton.SetDisabled(false)
	} else {
		p.DoneButton.SetDisabled(true)
	}
	if p.passwordsMatch() {
		p.CopyPassword.SetDisabled(false)
	} else {
		p.CopyPassword.SetDisabled(true)
	}
	return nil
}

func (p *EditPage) passwordsMatch() bool {
	return p.RepeatPassword.Value() == p.Password.Value()
}

func (p *EditPage) isReady() bool {
	return !p.isEmpty() &&
		len(p.Password.Value()) > 0 &&
		p.RepeatPassword.Value() == p.Password.Value() &&
		len(p.Username.Value()) > 0 &&
		len(p.Sitename.Value()) > 0
}

func (p *EditPage) isEmpty() bool {
	return p.Password.Value() == p.data.Password &&
		p.RepeatPassword.Value() == p.data.Password &&
		p.Username.Value() == p.data.Username &&
		p.Sitename.Value() == p.data.Sitename &&
		p.PasswordType.Value() == ""
}

func (p *EditPage) getTitle() string {
	if len(p.data.Filename) > 0 {
		return p.data.Filename
	} else if len(p.data.Sitename) > 0 {
		return p.data.Sitename
	} else {
		return "New Password"
	}
}
