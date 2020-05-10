// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"net"
	"net/url"
	"sort"
	"strings"
	"time"

	"golang.org/x/xerrors"
	"honnef.co/go/js/dom/v2"

	"github.com/bvk/past/msg"
)

const searchTemplate = `
<div>
	<div class="row header">
		<input class="cell-elastic" gotag="name:SearchBar input:OnSearchBar" type="text" placeholder="What are you looking for?"></input>
	</div>

	<hr/>

	<div class="content mw32em">
		<ul>
			<li gotag="name:SearchItemTemplate" style="display:none">
				<div class="row">
					<span class="cell-elastic" searchitem="name:PasswordName"></span>
					<button class="cell material-icons" searchitem="name:CopyButton click:OnCopyButton">content_copy</button>
					<button class="cell material-icons" searchitem="name:ViewButton click:OnViewButton">navigate_next</button>
				</div>
			</li>
		</ul>
	</div>

	<div class="row footer">
		<button class="cell material-icons" gotag="name:SettingsButton click:OnSettingsButton">tune</button>
		<div class="cell-elastic footer-status"></div>
		<button class="cell material-icons" gotag="name:AddButton click:OnAddButton">add</button>
	</div>
</div>
`

type SearchPage struct {
	ctl *Controller

	data *msg.ListFilesResponse

	usageMap map[string]int

	minListSize int
	maxListSize int

	root dom.Element

	items map[string]*SearchItem

	emptyItems []*SearchItem

	SearchBar          *dom.HTMLInputElement
	SearchItemTemplate *dom.HTMLLIElement

	SettingsButton, AddButton *dom.HTMLButtonElement
}

type SearchItem struct {
	page *SearchPage

	root dom.HTMLElement

	PasswordName *dom.HTMLSpanElement

	CopyButton, ViewButton *dom.HTMLButtonElement
}

func (i *SearchItem) OnViewButton(dom.Event) (status error) {
	req := msg.ViewFileRequest{
		Filename: i.PasswordName.TextContent(),
	}

	bs := StartBusy(i.page.root)
	resp, err := i.page.ctl.ViewFile(&req)
	StopBusy(bs)

	if err != nil {
		return i.page.ctl.setError(err)
	}
	page, err := NewViewPage(i.page.ctl, resp)
	if err != nil {
		return i.page.ctl.setError(err)
	}
	if err := i.page.ctl.ShowPage(page); err != nil {
		return err
	}
	return nil
}

func (i *SearchItem) OnCopyButton(dom.Event) (status error) {
	req := msg.ViewFileRequest{
		Filename: i.PasswordName.TextContent(),
	}

	bs := StartBusy(i.page.root)
	resp, err := i.page.ctl.ViewFile(&req)
	StopBusy(bs)

	if err != nil {
		return i.page.ctl.setError(err)
	}
	if err := i.page.ctl.CopyTimeout(resp.Password, 10*time.Second); err != nil {
		return err
	}
	i.page.usageMap[req.Filename] += 1
	return i.page.ctl.UpdateUseCount(i.page.usageMap)
}

func NewSearchPage(ctl *Controller, data *msg.ListFilesResponse, usageMap map[string]int) (*SearchPage, error) {
	// Add or remove password entries from the usageMap.
	newMap := make(map[string]int)
	for _, file := range data.Files {
		if c, ok := usageMap[file]; ok {
			newMap[file] = c
		} else {
			newMap[file] = 0
		}
	}

	p := &SearchPage{
		ctl:         ctl,
		data:        data,
		usageMap:    newMap,
		minListSize: 6,
		maxListSize: 8,
		items:       make(map[string]*SearchItem),
	}

	root, err := BindHTML(searchTemplate, "gotag", p)
	if err != nil {
		return nil, xerrors.Errorf("could not parse search page template: %w", err)
	}
	p.root = root

	for _, file := range data.Files {
		clone := mustCloneHTMLElement(p.SearchItemTemplate, true)

		item := &SearchItem{page: p, root: clone}
		if err := BindTarget(clone, "searchitem", item); err != nil {
			return nil, p.ctl.setError(err)
		}
		item.PasswordName.SetTextContent(file)
		p.items[file] = item
	}

	for i := 0; i < p.minListSize; i++ {
		clone := mustCloneHTMLElement(p.SearchItemTemplate, true)

		item := &SearchItem{page: p, root: clone}
		if err := BindTarget(clone, "searchitem", item); err != nil {
			return nil, p.ctl.setError(err)
		}
		item.CopyButton.SetDisabled(true)
		item.ViewButton.SetDisabled(true)
		p.emptyItems = append(p.emptyItems, item)
	}
	return p, nil
}

func ShowSearchPage(ctl *Controller) (*SearchPage, error) {
	resp, err := ctl.ListFiles()
	if err != nil {
		return nil, err
	}
	page, err := NewSearchPage(ctl, resp, ctl.UseCountMap())
	if err != nil {
		return nil, err
	}
	if err := ctl.ShowPage(page); err != nil {
		return nil, err
	}
	return page, nil
}

func (p *SearchPage) RootDiv() dom.Element {
	return p.root
}

func (p *SearchPage) RefreshDisplay() error {
	var files []string
	if s := p.SearchBar.Value(); len(s) > 0 {
		files = p.searchPasswordFiles(s)
	} else {
		files = p.orderPasswordFiles()
	}

	// Remove all previous entries first.
	for _, item := range p.items {
		if p := item.root.ParentNode(); p != nil {
			p.RemoveChild(item.root)
		}
	}
	for _, item := range p.emptyItems {
		if p := item.root.ParentNode(); p != nil {
			p.RemoveChild(item.root)
		}
	}

	for i := 0; i < p.maxListSize && i < len(files); i++ {
		file := files[i]
		item := p.items[file]

		item.root.Style().RemoveProperty("display")
		p.SearchItemTemplate.ParentNode().InsertBefore(item.root, p.SearchItemTemplate)
	}

	for i := len(files); i < p.minListSize; i++ {
		item := p.emptyItems[i]
		item.root.Style().RemoveProperty("display")
		p.SearchItemTemplate.ParentNode().InsertBefore(item.root, p.SearchItemTemplate)
	}

	return nil
}

func (p *SearchPage) OnSearchBar(dom.Event) (status error) {
	return p.RefreshDisplay()
}

func (p *SearchPage) OnAddButton(dom.Event) (status error) {
	page, err := NewEditPage(p.ctl, new(msg.ViewFileResponse))
	if err != nil {
		return err
	}
	if err := p.ctl.ShowPage(page); err != nil {
		return err
	}
	return nil
}

func (p *SearchPage) OnSettingsButton(dom.Event) (status error) {
	if _, err := ShowSettingsPage(p.ctl); err != nil {
		return err
	}
	return nil
}

func (p *SearchPage) searchPasswordFiles(search string) []string {
	hosts, _ := p.getActiveTabHostSuffixes()
	files := p.sortPasswordFiles()

	// We only want to show search-maches and hostname-matches, nothing else.

	var selected, rest []string
	for _, file := range files {
		if strings.Contains(file, search) {
			selected = append(selected, file)
		} else {
			rest = append(rest, file)
		}
	}
	for _, file := range rest {
		for _, host := range hosts {
			if strings.Contains(file, host) {
				selected = append(selected, file)
				break
			}
		}
	}
	return selected
}

func (p *SearchPage) orderPasswordFiles() []string {
	hosts, _ := p.getActiveTabHostSuffixes()
	files := p.sortPasswordFiles()

	// We want to show hostname-matches first followed by all the rest in
	// most-used order.

	var selected, rest []string
	for _, file := range files {
		hostMatch := false
		for _, host := range hosts {
			if strings.Contains(file, host) {
				hostMatch = true
				selected = append(selected, file)
				break
			}
		}
		if !hostMatch {
			rest = append(rest, file)
		}
	}
	return append(selected, rest...)
}

func (p *SearchPage) sortPasswordFiles() []string {
	files := append([]string{}, p.data.Files...)
	sort.Slice(files, func(i, j int) bool {
		return p.usageMap[files[i]] > p.usageMap[files[j]]
	})
	return files
}

func (p *SearchPage) getActiveTabHostSuffixes() ([]string, error) {
	tabAddr := p.ctl.ActiveTabAddress()
	if len(tabAddr) == 0 {
		return nil, nil
	}
	tabURL, err := url.Parse(tabAddr)
	if err != nil {
		return nil, err
	}

	// If hostname is in IP or IP:Port form, we should not split the hostname
	// into suffixes.
	hostname := tabURL.Host
	if strings.ContainsRune(tabURL.Host, ':') {
		if host, _, err := net.SplitHostPort(tabURL.Host); err == nil {
			if ip := net.ParseIP(host); ip != nil {
				return []string{tabURL.Host}, nil
			}
			hostname = host
		}
	}

	var suffixes []string
	words := strings.Split(hostname, ".")
	for ii := 0; ii < len(words)-1; ii++ {
		suffixes = append(suffixes, strings.Join(words[ii:], "."))
	}

	return suffixes, nil
}
