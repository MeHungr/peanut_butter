BINARY_DIR := ./bin
GO := go

PKG_SERVER := ./cmd/server
PKG_AGENT := ./cmd/agent
PKG_CLI := ./cmd/cli

HOSTOS := $(shell go env GOOS)
HOSTARCH := $(shell go env GOARCH)

# Release ldflags: strip symbol & DWARF, and trim paths
LDFLAGS := -s -w

# Default build type (release)
.PHONY: all
all: clean linux-x64 linux-arm mac freebsd windows install

# Default install prefix (Linux/macOS)
PREFIX ?= $(HOME)/.local

# -------- TLS Certs --------

.PHONY: build-certs
build-certs: ca gen-server rotate-cert

.PHONY: ca
ca:
	@openssl genrsa -out ca.key 4096
	@openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt -subj "/CN=PNUTBR"

.PHONY: gen-server
gen-server:
	@openssl genrsa -out server.key 2048
	@openssl req -new -key server.key -out server.csr -subj "/CN=server.local"
	@openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365

.PHONY: rotate-cert
rotate-cert:

# -------- Release builds (stripped) --------

# Linux x64 install
.PHONY: linux-x64
linux-x64:
	@mkdir -p $(BINARY_DIR)/linux_amd64
	GOOS=linux GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/linux_amd64/pbserver $(PKG_SERVER)
	GOOS=linux GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/linux_amd64/pbagent $(PKG_AGENT)
	GOOS=linux GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/linux_amd64/pbctl $(PKG_CLI)

# Linux ARM install
.PHONY: linux-arm
linux-arm:
	@mkdir -p $(BINARY_DIR)/linux_arm64
	GOOS=linux GOARCH=arm64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/linux_arm64/pbserver $(PKG_SERVER)
	GOOS=linux GOARCH=arm64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/linux_arm64/pbagent $(PKG_AGENT)
	GOOS=linux GOARCH=arm64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/linux_arm64/pbctl $(PKG_CLI)

# Mac install
.PHONY: mac
mac:
	@mkdir -p $(BINARY_DIR)/darwin_amd64
	GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/darwin_amd64/pbserver $(PKG_SERVER)
	GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/darwin_amd64/pbagent $(PKG_AGENT)
	GOOS=darwin GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/darwin_amd64/pbctl $(PKG_CLI)

# FreeBSD install
.PHONY: freebsd
freebsd:
	@mkdir -p $(BINARY_DIR)/darwin_amd64
	GOOS=freebsd GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/freebsd_amd64/pbserver $(PKG_SERVER)
	GOOS=freebsd GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/freebsd_amd64/pbagent $(PKG_AGENT)
	GOOS=freebsd GOARCH=amd64 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/freebsd_amd64/pbctl $(PKG_CLI)

# Windows install
.PHONY: windows
windows:
ifeq ($(OS),Windows_NT)
	@mkdir -p $(BINARY_DIR)/windows_amd64
	GOOS=windows GOARCH=amd64\
		$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/windows_amd64/pbserver.exe $(PKG_SERVER)
	GOOS=windows GOARCH=amd64\
		$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/windows_amd64/pbagent.exe $(PKG_AGENT)
	GOOS=windows GOARCH=amd64\
		$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/windows_amd64/pbctl.exe $(PKG_CLI)
else
	@command -v x86_64-w64-mingw32-gcc >/dev/null 2>&1 || \
		{ echo "Error: mingw-w64 not installed (need x86_64-w64-mingw32-gcc)"; exit 1; }
	@mkdir -p $(BINARY_DIR)/windows_amd64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
		$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/windows_amd64/pbserver.exe $(PKG_SERVER)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
		$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/windows_amd64/pbagent.exe $(PKG_AGENT)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc \
		$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/windows_amd64/pbctl.exe $(PKG_CLI)
endif

# -------- Install CLI on Linux/macOS --------
.PHONY: install
install:
	@echo "Installing pbctl to $(PREFIX)/bin..."
	install -d $(PREFIX)/bin
	install $(BINARY_DIR)/$(HOSTOS)_$(HOSTARCH)/pbctl $(PREFIX)/bin/pbctl

	@echo "Installing shell tab completions..."
	# Bash
	install -d $(PREFIX)/share/bash-completion/completions
	$(BINARY_DIR)/$(HOSTOS)_$(HOSTARCH)/pbctl completion bash > $(PREFIX)/share/bash-completion/completions/pbctl
	# Zsh
	install -d $(PREFIX)/share/zsh/site-functions
	$(BINARY_DIR)/$(HOSTOS)_$(HOSTARCH)/pbctl completion zsh > $(PREFIX)/share/zsh/site-functions/_pbctl
	# Fish
	install -d $(PREFIX)/share/fish/vendor_completions.d
	$(BINARY_DIR)/$(HOSTOS)_$(HOSTARCH)/pbctl completion fish > $(PREFIX)/share/fish/vendor_completions.d/pbctl.fish

	@echo "pbctl installed successfully. Restart your shell or source the completion file."

# -------- Clean --------
.PHONY: clean
clean:
	-rm -rf $(BINARY_DIR)
	-mkdir -p $(BINARY_DIR)
