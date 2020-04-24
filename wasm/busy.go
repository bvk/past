// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"fmt"
	"time"

	"honnef.co/go/js/dom/v2"
)

type busyState struct {
	stopCh chan struct{}

	disabledItems []EnableDisable
}

func StartBusy(root dom.Element) *busyState {
	return &busyState{
		stopCh:        showBusy(root),
		disabledItems: disableElems(root),
	}
}

func StopBusy(bs *busyState) {
	close(bs.stopCh)
	enableElems(bs.disabledItems)
}

func showBusy(root dom.Element) chan struct{} {
	var hrs []*dom.HTMLHRElement
	for _, e := range root.GetElementsByTagName("hr") {
		hrs = append(hrs, e.(*dom.HTMLHRElement))
	}

	id := make(chan struct{})
	go func() {
		dir := "left"
		leftMargin, rightMargin := 25, 25

		for {
			select {
			case <-id:
				for _, hr := range hrs {
					hr.Style().RemoveProperty("margin-left")
					hr.Style().RemoveProperty("margin-right")
				}
				return
			case <-time.After(20 * time.Millisecond):
				for _, hr := range hrs {
					hr.Style().SetProperty("margin-left", fmt.Sprintf("%d%%", leftMargin), "")
					hr.Style().SetProperty("margin-right", fmt.Sprintf("%d%%", rightMargin), "")
				}
			}

			if dir == "left" {
				leftMargin++
				rightMargin--
			} else {
				leftMargin--
				rightMargin++
			}

			if rightMargin == 0 {
				dir = "right"
			}
			if leftMargin == 0 {
				dir = "left"
			}
		}
	}()
	return id
}

type EnableDisable interface {
	Disabled() bool
	SetDisabled(bool)
}

func disableElems(root dom.Element) []EnableDisable {
	var items []EnableDisable

	var walk func(dom.Node)
	walk = func(n dom.Node) {
		if item, ok := n.(EnableDisable); ok {
			if !item.Disabled() {
				items = append(items, item)
				item.SetDisabled(true)
			}
		}
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			walk(c)
		}
	}
	walk(root)

	return items
}

func enableElems(elems []EnableDisable) {
	for _, elem := range elems {
		elem.SetDisabled(false)
	}
}
