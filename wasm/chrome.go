// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"log"
	"os"
	"syscall/js"

	"golang.org/x/xerrors"
)

func SendNativeMessage(name string, object js.Value) js.Value {
	resp := make(chan js.Value, 1)
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cb.Release()
		if len(args) != 0 {
			resp <- args[0]
		} else {
			resp <- js.Undefined()
		}
		return nil
	})
	// FIXME: We should use js.Value types' Equal method with go1.13+
	chrome := js.Global().Get("chrome")
	if chrome == js.Undefined() || chrome == js.Null() {
		return js.Undefined()
	}
	runtime := chrome.Get("runtime")
	if runtime == js.Undefined() || runtime == js.Null() {
		return js.Undefined()
	}
	runtime.Call("sendNativeMessage", name, object, cb)
	return <-resp
}

func ActiveTab() js.Value {
	resp := make(chan js.Value, 1)
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cb.Release()
		if len(args) != 0 {
			resp <- args[0]
		} else {
			resp <- js.Undefined()
		}
		return nil
	})
	// FIXME: We should use js.Value types' Equal method with go1.13+
	chrome := js.Global().Get("chrome")
	if chrome == js.Undefined() || chrome == js.Null() {
		return js.Undefined()
	}
	tabs := chrome.Get("tabs")
	if tabs == js.Undefined() || tabs == js.Null() {
		return js.Undefined()
	}
	tabs.Call("getSelected", js.Null(), cb)
	return <-resp
}

func GetBackgroundPage() js.Value {
	resp := make(chan js.Value, 1)
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cb.Release()
		if len(args) != 0 {
			resp <- args[0]
		} else {
			resp <- js.Undefined()
		}
		return nil
	})
	// FIXME: We should use js.Value types' Equal method with go1.13+
	chrome := js.Global().Get("chrome")
	if chrome == js.Undefined() || chrome == js.Null() {
		return js.Undefined()
	}
	runtime := chrome.Get("runtime")
	if runtime == js.Undefined() || runtime == js.Null() {
		return js.Undefined()
	}
	runtime.Call("getBackgroundPage", cb)
	return <-resp
}

func SetContextMenuHandler(callback func(info, tab js.Value)) js.Value {
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cb.Release()
		if len(args) < 2 {
			log.Println("error: at least two arguments are unexpected in the callback")
			return nil
		}
		callback(args[0], args[1])
		return nil
	})

	chrome := js.Global().Get("chrome")
	if chrome == js.Undefined() || chrome == js.Null() {
		return js.Undefined()
	}
	menus := chrome.Get("contextMenus")
	if menus == js.Undefined() || menus == js.Null() {
		return js.Undefined()
	}
	clicked := chrome.Get("onClicked")
	if clicked == js.Undefined() || clicked == js.Null() {
		return js.Undefined()
	}
	return clicked.Call("addListener", cb)
}

func AddContextMenu(id, name string, menuContexts, urlMatches []string) js.Value {
	var matches []interface{}
	for _, match := range urlMatches {
		matches = append(matches, js.ValueOf(match))
	}
	var contexts []interface{}
	for _, context := range menuContexts {
		contexts = append(contexts, js.ValueOf(context))
	}

	menudata := map[string]interface{}{
		"id":                  js.ValueOf(id),
		"visible":             js.ValueOf(true),
		"enabled":             js.ValueOf(true),
		"documentUrlPatterns": js.ValueOf(matches),
		"title":               js.ValueOf(name),
		"contexts":            js.ValueOf(contexts),
	}
	chrome := js.Global().Get("chrome")
	if chrome == js.Undefined() || chrome == js.Null() {
		return js.Undefined()
	}
	menus := chrome.Get("contextMenus")
	if menus == js.Undefined() || menus == js.Null() {
		return js.Undefined()
	}
	return menus.Call("create", js.ValueOf(menudata))
}

func RemoveContextMenu(id string) {
	resp := make(chan js.Value, 1)
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cb.Release()
		if len(args) != 0 {
			resp <- args[0]
		} else {
			resp <- js.Undefined()
		}
		return nil
	})
	chrome := js.Global().Get("chrome")
	if chrome == js.Undefined() || chrome == js.Null() {
		return
	}
	menus := chrome.Get("contextMenus")
	if menus == js.Undefined() || menus == js.Null() {
		return
	}
	menus.Call("remove", js.ValueOf(id), cb)
	<-resp
}

func RemoveAllContextMenus() {
	resp := make(chan js.Value, 1)
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cb.Release()
		if len(args) != 0 {
			resp <- args[0]
		} else {
			resp <- js.Undefined()
		}
		return nil
	})
	chrome := js.Global().Get("chrome")
	if chrome == js.Undefined() || chrome == js.Null() {
		return
	}
	menus := chrome.Get("contextMenus")
	if menus == js.Undefined() || menus == js.Null() {
		return
	}
	menus.Call("removeAll", cb)
	<-resp
}

func GetItem(key string) js.Value {
	resp := make(chan js.Value, 1)
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cb.Release()
		if len(args) != 0 {
			resp <- args[0]
		} else {
			resp <- js.Undefined()
		}
		return nil
	})
	chrome := js.Global().Get("chrome")
	if chrome == js.Undefined() || chrome == js.Null() {
		return js.Undefined()
	}
	storage := chrome.Get("storage")
	if storage == js.Undefined() || storage == js.Null() {
		return js.Undefined()
	}
	local := storage.Get("local")
	if storage == js.Undefined() || storage == js.Null() {
		return js.Undefined()
	}
	local.Call("get", js.ValueOf(key), cb)
	return <-resp
}

func SetItem(key, val string) error {
	resp := make(chan js.Value, 1)
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		cb.Release()
		if len(args) != 0 {
			resp <- args[0]
		} else {
			resp <- js.Undefined()
		}
		return nil
	})
	chrome := js.Global().Get("chrome")
	if chrome == js.Undefined() || chrome == js.Null() {
		return xerrors.Errorf("could not get reference to chrome: %w", os.ErrInvalid)
	}
	storage := chrome.Get("storage")
	if storage == js.Undefined() || storage == js.Null() {
		return xerrors.Errorf("could not get reference to chrome.storage: %w", os.ErrInvalid)
	}
	local := storage.Get("local")
	if storage == js.Undefined() || storage == js.Null() {
		return xerrors.Errorf("could not get reference to chrome.storage.local: %w", os.ErrInvalid)
	}
	item := map[string]interface{}{
		key: val,
	}
	local.Call("set", js.ValueOf(item), cb)
	<-resp
	return nil
}
