# Copyright (c) 2020 BVK Chaitanya

export GO ?= go
export GOBIN ?= $(CURDIR)
export PATH := $(PATH):$(HOME)/go/bin

VERSION := $(shell git describe --tags --always --dirty)

BUILD_DATE := $(shell date "+%Y%m%d")
BUILD_TIME := $(shell date "+%H%M%S")
BUILD_DIR ?= $(CURDIR)/build/$(VERSION)-$(BUILD_DATE)-$(BUILD_TIME)

.PHONY: all
all: go-generate go-all;

.PHONY: check
check: all
	$(MAKE) go-test

.PHONY: clean
clean:
	git clean -f -X

.PHONY: bash
bash:
	@echo \#
	@echo \# Interactive BASH SHELL in the build environment.
	@echo \#
	bash -li

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

.PHONY: go-wasm-all
go-wasm-all:
	GOOS=js GOARCH=wasm $(GO) build -o $(CURDIR)/extension/main.wasm github.com/bvk/past/wasm
	cp $(shell $(GO) env GOROOT)/misc/wasm/wasm_exec.js $(CURDIR)/extension/wasm_exec.js

.PHONY: go-generate
go-generate: go-wasm-all
	$(GO) generate ./...

.PHONY: go-all
go-all: $(BUILD_DIR)
	$(GO) build github.com/bvk/past/cmd/past

.PHONY: go-test
go-test: $(BUILD_DIR) go-all
	TMPDIR=$(BUILD_DIR) $(GO) test -count=1 $(if $(GOTESTFLAGS),-v) $(GOTESTFLAGS) ./...

# Include workspace local make rules if any.
-include $(CURDIR)/../Makefile.custom
