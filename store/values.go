// Copyright (c) 2020 BVK Chaitanya

package store

import (
	"bytes"
	"regexp"
	"sort"
	"strings"
)

var kvRe = regexp.MustCompile("(?s:\n(?:[^\t]))")

type Values struct {
	m map[string]string
}

// NewValues creates a key-value set from the input data. Input data can also
// be nil in which case an empty key-value set is created.
//
// Keys and values are case sensitive, but cannot begin or end with whitespace.
// Keys must not include ':' character cause it is used to separate key and
// value data in the raw bytes form.
func NewValues(data []byte) *Values {
	vs := &Values{m: make(map[string]string)}

	var items []string
	for i := 0; len(data) > 0; i++ {
		end := kvRe.Find(data)
		if end == nil {
			items = append(items, string(data))
			break
		}
		p := bytes.Index(data, end)
		items = append(items, string(data[:p]))
		data = data[p+1:]
	}

	for _, item := range items {
		if len(item) == 0 {
			continue
		}
		kv := strings.Replace(item, "\n\t", "\n", -1)
		index := strings.IndexRune(kv, ':')
		if index == -1 {
			vs.m[strings.TrimSpace(kv)] = ""
			continue
		}
		k, v := kv[:index], kv[index+1:]
		if len(k) == 0 && len(v) == 0 {
			continue
		}
		vs.m[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}

	return vs
}

// Bytes serializes key-value pairs into raw bytes.
//
// Newline characters in the keys and values are translated into `\n\t`
// sequence, so that key-value pairs can easily interpreted in the editors.
// Since keys do not begin with a whitespace and always start on a newline,
// they can be used to identify the end-of-value.
func (vs *Values) Bytes() []byte {
	var keys []string
	for k := range vs.m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf bytes.Buffer
	for _, key := range keys {
		value := vs.m[key]
		buf.WriteString(strings.Replace(key, "\n", "\n\t", -1))
		buf.WriteRune(':')
		buf.WriteString(strings.Replace(value, "\n", "\n\t", -1))
		buf.WriteRune('\n')
	}
	return buf.Bytes()
}

// Get returns the value associated with the key. Whitespace around the keys is
// always trimmed. Keys with ':' character are invalid.
func (vs *Values) Get(key string) string {
	if strings.ContainsRune(key, ':') {
		return ""
	}
	return vs.m[strings.TrimSpace(key)]
}

// Set adds or updates a value to a key. Whitespace around the keys and values
// is always trimmed.
func (vs *Values) Set(key, value string) {
	if !strings.ContainsRune(key, ':') {
		vs.m[strings.TrimSpace(key)] = value
	}
}

// Del removes a key value pair from the values.
func (vs *Values) Del(key string) {
	if !strings.ContainsRune(key, ':') {
		delete(vs.m, strings.TrimSpace(key))
	}
}

// Len returns the number of key-value pairs.
func (vs *Values) Len() int {
	return len(vs.m)
}

// GetUsernames is a helper function to identify a key that represents an
// username. One of the "username", "user" or "login" keys (in case-insensitive
// form) identify an username.
func GetUsernames(vs *Values) []string {
	var users []string
	for k, v := range vs.m {
		kk := strings.ToLower(k)
		if kk == "username" || kk == "user" || kk == "login" {
			users = append(users, v)
		}
	}
	return users
}
