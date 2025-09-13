BINARY_DIR := ./bin
DIST_DIR := ./dist
GO := go

PKG_SERVER := ./cmd/server
PKG_AGENT := ./cmd/agent
PKG_CLI := ./cmd/cli

# Release ldflags: strip symbol & DWARF, and trim paths
LDFLAGS := -s -w

# Default build type (dev)
.PHONY: all
	all: build

# DEV builds (keeps debug info)
.PHONY: build
	build: build-server build-agent build-cli

.PHONY: build-server
	build-server:
	@mkdir -p $(BINARY_DIR)
	$(GO) build -o $(BINARY_DIR)/server $(PKG_SERVER)

.PHONY: build-agent
build-agent:
	@mkdir -p $(BINARY_DIR)
	$(GO) build -o $(BINARY_DIR)/agent $(PKG_AGENT)

.PHONY: build-cli
build-cli:
	@mkdir -p $(BINARY_DIR)
	$(GO) build -o $(BINARY_DIR)/pbctl $(PKG_CLI)

# Release: cross-compile for multiple platforms and create dist/
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 windows/amd64

.PHONY: release
release: clean-dist
	@echo "Building release (VERSION=$(VERSION))..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
	    os=$${platform%/*}; arch=$${platform#*/}; \
	    outdir=$(DIST_DIR)/$${os}_$${arch}; mkdir -p $$outdir; \
	    echo "Building server for $$platform..."; \
	    env GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $$outdir/server $(PKG_SERVER); \
	    echo "Building agent for $$platform..."; \
	    env GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $$outdir/agent $(PKG_AGENT); \
	    echo "Building cli for $$platform..."; \
	    env GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 $(GO) build -trimpath -ldflags="$(LDFLAGS)" -o $$outdir/pbctl $(PKG_CLI); \
	done
	@echo "Release builds done. Dist dir: $(DIST_DIR)"

.PHONY: clean-dist
clean-dist:
	-rm -rf $(DIST_DIR)
	-mkdir -p $(DIST_DIR)
