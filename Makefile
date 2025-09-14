BINARY_DIR := ./bin
GO := go

PKG_SERVER := ./cmd/server
PKG_AGENT := ./cmd/agent
PKG_CLI := ./cmd/cli

# Release ldflags: strip symbol & DWARF, and trim paths
LDFLAGS := -s -w

# Default build type (release)
.PHONY: all
all: release

# Default install prefix (Linux/macOS)
PREFIX ?= $(HOME)/.local

# Installs pbctl on linux
.PHONY: install
install: release-cli
	@echo "Installing pbctl to $(PREFIX)/bin..."
	install -d $(PREFIX)/bin
	install $(BINARY_DIR)/pbctl $(PREFIX)/bin/pbctl

	@echo "Installing shell tab completions..."
	# Bash
	install -d $(PREFIX)/share/bash-completion/completions
	$(BINARY_DIR)/pbctl completion bash > $(PREFIX)/share/bash-completion/completions/pbctl
	# Zsh
	install -d $(PREFIX)/share/zsh/site-functions
	$(BINARY_DIR)/pbctl completion zsh > $(PREFIX)/share/zsh/site-functions/_pbctl
	# Fish
	install -d $(PREFIX)/share/fish/vendor_completions.d
	$(BINARY_DIR)/pbctl completion fish > $(PREFIX)/share/fish/vendor_completions.d/pbctl.fish

	@echo "pbctl installed successfully. Restart your shell or source the completion file."

# DEV builds (keeps debug info)
.PHONY: dev
dev: clean build-server build-agent build-cli

.PHONY: build-server
build-server:
	@mkdir -p $(BINARY_DIR)
	$(GO) build -o $(BINARY_DIR)/pbserver $(PKG_SERVER)

.PHONY: build-agent
build-agent:
	@mkdir -p $(BINARY_DIR)
	$(GO) build -o $(BINARY_DIR)/pbagent $(PKG_AGENT)

.PHONY: build-cli
build-cli:
	@mkdir -p $(BINARY_DIR)
	$(GO) build -o $(BINARY_DIR)/pbctl $(PKG_CLI)

# Release: cross-compile for multiple platforms and create dist/
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 windows/amd64

.PHONY: release
release: clean
	@echo "Building release"
	@mkdir -p $(BINARY_DIR)
	@for platform in $(PLATFORMS); do \
	    os=$${platform%/*}; arch=$${platform#*/}; \
	    outdir=$(BINARY_DIR)/$${os}_$${arch}; mkdir -p $$outdir; \
	    echo "Building server for $$platform..."; \
	    env GOOS=$$os GOARCH=$$arch $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $$outdir/pbserver $(PKG_SERVER); \
	    echo "Building agent for $$platform..."; \
	    env GOOS=$$os GOARCH=$$arch $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $$outdir/pbagent $(PKG_AGENT); \
	    echo "Building cli for $$platform..."; \
	    env GOOS=$$os GOARCH=$$arch $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $$outdir/pbctl $(PKG_CLI); \
	done
	@echo "Release builds done. Dist dir: $(BINARY_DIR)"

.PHONY: release-cli
release-cli: clean
	@echo "Building pbctl (release version)"
	@mkdir -p $(BINARY_DIR)
	$(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $(BINARY_DIR)/pbctl $(PKG_CLI)
	@echo "Done building pbctl."

.PHONY: clean
clean:
	-rm -rf $(BINARY_DIR)
	-mkdir -p $(BINARY_DIR)
