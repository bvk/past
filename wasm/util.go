// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"honnef.co/go/js/dom/v2"

	"github.com/bvk/past"
)

func getButtonsByClassName(e dom.Element, name string) []*dom.HTMLButtonElement {
	var bs []*dom.HTMLButtonElement
	for _, e := range e.GetElementsByClassName(name) {
		if b, ok := e.(*dom.HTMLButtonElement); ok {
			bs = append(bs, b)
		}
	}
	return bs
}

func mustAtoi(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		log.Panic(err)
	}
	return v
}

func mustCloneHTMLElement(orig dom.HTMLElement, deep bool) dom.HTMLElement {
	dup := orig.CloneNode(deep)
	elem, ok := dup.(dom.HTMLElement)
	if !ok {
		log.Panicf("could not convert cloned node to an element")
	}
	return elem
}

func rotateKeyID(keyid dom.Element, key *past.PublicKeyData) {
	switch v := keyid.TextContent(); v {
	case key.UserName:
		keyid.SetTextContent(key.UserEmail)
	case key.UserEmail:
		keyid.SetTextContent(key.KeyFingerprint)
	case key.KeyFingerprint:
		keyid.SetTextContent(key.UserName)
	}
}

func pwgen(charset string, length int) string {
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteByte(charset[rand.Intn(len(charset))])
	}
	return b.String()
}

func marshalIndentString(v interface{}, prefix, indent string) (string, error) {
	data, err := json.MarshalIndent(v, prefix, indent)
	return string(data), err
}
