// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"log"
	"os"
	"reflect"
	"strings"

	"golang.org/x/xerrors"
	"honnef.co/go/js/dom/v2"
)

func ParseHTML(div string) (dom.Element, error) {
	w := dom.GetWindow()
	d := w.Document()
	t := d.CreateElement("div")
	t.SetInnerHTML(div)
	for c := t.FirstChild(); c != nil; c = c.NextSibling() {
		if ce, ok := c.(dom.Element); ok {
			n := ce.CloneNode(true)
			if ne, ok := n.(dom.Element); ok {
				return ne, nil
			}
			return nil, xerrors.Errorf("clone did not return %T an element: %w", n, os.ErrInvalid)
		}
	}
	return nil, xerrors.Errorf("could not find any html element: %w", os.ErrInvalid)
}

func BindHTML(html, attr string, target interface{}) (dom.Element, error) {
	root, err := ParseHTML(html)
	if err != nil {
		return nil, err
	}
	if err := BindTarget(root, attr, target); err != nil {
		return nil, err
	}
	return root, nil
}

func BindTarget(root dom.Element, attr string, target interface{}) error {
	var tagTypeMap = map[string]reflect.Type{
		"button":   reflect.TypeOf(&dom.HTMLButtonElement{}),
		"li":       reflect.TypeOf(&dom.HTMLLIElement{}),
		"span":     reflect.TypeOf(&dom.HTMLSpanElement{}),
		"input":    reflect.TypeOf(&dom.HTMLInputElement{}),
		"div":      reflect.TypeOf(&dom.HTMLDivElement{}),
		"select":   reflect.TypeOf(&dom.HTMLSelectElement{}),
		"textarea": reflect.TypeOf(&dom.HTMLTextAreaElement{}),
	}

	var events = []string{"click", "input", "change", "wheel"}

	var attrKeys = append([]string{"name"}, events...)

	type Element struct {
		node dom.Element

		name, click, input, wheel, change string

		keyValueMap map[string]string
	}

	nameMap := make(map[string]*Element)
	cb := func(node dom.Element) error {
		attrValue := node.GetAttribute(attr)
		if len(attrValue) == 0 {
			return nil
		}

		tag := strings.ToLower(node.TagName())
		if _, ok := tagTypeMap[tag]; !ok {
			return xerrors.Errorf("attribute found on unsupported tag %q: %w", tag, os.ErrInvalid)
		}

		keyValueMap := make(map[string]string)
		for _, word := range strings.Fields(attrValue) {
			key, value := "", ""
			for _, k := range attrKeys {
				if strings.HasPrefix(word, k+":") {
					key, value = k, strings.TrimPrefix(word, k+":")
					break
				}
			}

			if _, ok := keyValueMap[key]; ok {
				return xerrors.Errorf("key %q is duplicated: %w", os.ErrInvalid)
			}

			if len(value) == 0 {
				continue
			}

			keyValueMap[key] = value
		}

		elem := &Element{
			node:        node,
			name:        keyValueMap["name"],
			click:       keyValueMap["click"],
			input:       keyValueMap["input"],
			change:      keyValueMap["change"],
			wheel:       keyValueMap["wheel"],
			keyValueMap: keyValueMap,
		}

		if len(elem.name) == 0 {
			return xerrors.Errorf("attribute has no 'name' key: %w", os.ErrInvalid)
		}

		if _, ok := nameMap[elem.name]; ok {
			return xerrors.Errorf("multiple elements has same 'name:%s' binding: %w", elem.name, os.ErrInvalid)
		}

		nameMap[elem.name] = elem
		return nil
	}

	var walk func(dom.Node) error
	walk = func(n dom.Node) error {
		if n.NodeType() == 1 /* Element */ {
			if err := cb(n.(dom.Element)); err != nil {
				return xerrors.Errorf("could not identify/parse attribute %q: %w", attr, err)
			}
		}
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if err := walk(c); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(root); err != nil {
		return xerrors.Errorf("couldn't complete element walk: %w", err)
	}

	// Make sure v refers to the struct, not the pointer.
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return xerrors.Errorf("target must be a struct or pointer to a struct: %w", os.ErrInvalid)
	}

	// Event methods can be in one of the two following types.
	var method1 func(dom.Event)
	var method2 func(dom.Event) error
	method1Type := reflect.TypeOf(method1)
	method2Type := reflect.TypeOf(method2)
	methodTypes := []string{method1Type.String(), method2Type.String()}

	var zero reflect.Value
	for name, elem := range nameMap {
		// Check that target has a field the given name.
		field := v.FieldByName(name)
		if field == zero {
			return xerrors.Errorf("target has no field with name %q: %w", name, os.ErrInvalid)
		}
		// Check that target's field has the expected HTML element type.
		tag := strings.ToLower(elem.node.TagName())
		if t := tagTypeMap[tag]; t != field.Type() {
			return xerrors.Errorf("target field %q type must be %q: %w", name, t.String(), os.ErrInvalid)
		}
		field.Set(reflect.ValueOf(elem.node))

		for _, event := range events {
			methodName, ok := elem.keyValueMap[event]
			if !ok || len(methodName) == 0 {
				continue
			}

			// Reset v to it's original type, which may be a pointer.
			v := reflect.ValueOf(target)
			method := v.MethodByName(methodName)
			if method == zero {
				return xerrors.Errorf("target has no method with name %q: %w", methodName, os.ErrInvalid)
			}

			// Check that target has event handler methods with one of the supported
			// types.
			if method.Type() == method1Type {
				elem.node.AddEventListener(event, true, func(ev dom.Event) {
					go method.Call([]reflect.Value{reflect.ValueOf(ev)})
				})
				continue
			}

			if method.Type() == method2Type {
				elem.node.AddEventListener(event, true, func(ev dom.Event) {
					go func() {
						rs := method.Call([]reflect.Value{reflect.ValueOf(ev)})
						if !rs[0].IsNil() {
							log.Printf("error: %v", rs[0].Interface())
						}
					}()
				})
				continue
			}

			return xerrors.Errorf("target method %q type must be one of %v types: %w", methodName, methodTypes, os.ErrInvalid)
		}
	}
	return nil
}
