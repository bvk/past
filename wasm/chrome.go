// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"syscall/js"
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
