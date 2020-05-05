// Copyright (c) 2020 BVK Chaitanya

package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"path"

	"golang.org/x/xerrors"

	"github.com/bvk/past/msg"
)

type ServerBackend struct {
	addr string
}

func NewServerBackend(href string) (*ServerBackend, error) {
	addrURL, err := url.Parse(href)
	if err != nil {
		return nil, err
	}
	addrURL.Path = path.Join(addrURL.Path, "/api")
	s := &ServerBackend{
		addr: addrURL.String(),
	}
	return s, nil
}

func (s *ServerBackend) Call(req *msg.Request) (*msg.Response, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(req); err != nil {
		return nil, err
	}
	resp, err := http.Post(s.addr, "application/json", &b)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	response := new(msg.Response)
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return nil, err
	}

	if response.ViewFile == nil {
		log.Println(marshalIndentString(response, "", "  "))
	} else {
		password := response.ViewFile.Password
		response.ViewFile.Password = "*"
		log.Println(marshalIndentString(response, "", "  "))
		response.ViewFile.Password = password
	}

	if len(response.Status) > 0 {
		return response, xerrors.New(response.Status)
	}
	return response, nil
}

func (s *ServerBackend) ActiveTabAddr() string {
	return "" // Server address doesn't qualify as the search string.
}
