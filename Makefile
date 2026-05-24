.PHONY: all build clean test vet cross cross-all router-builds

APP := lucx-core
OUT_DIR := build
LDFLAGS := -s -w
GO := go
GOENV := CGO_ENABLED=0 PATH=$(HOME)/.local/go/bin:$$PATH GOTOOLCHAIN=auto

all: test build

build:
	$(GOENV) $(GO) build -ldflags="$(LDFLAGS)" -o $(OUT_DIR)/$(APP) ./cmd/$(APP)/

test:
	$(GOENV) $(GO) test ./... -count=1

vet:
	$(GOENV) $(GO) vet ./...

clean:
	rm -rf $(OUT_DIR)

# Single target cross-compile
cross: cross-amd64 cross-arm64 cross-mipsle cross-armv7
	@echo "=== All cross-compilation targets built ==="

cross-amd64:
	@mkdir -p $(OUT_DIR)
	GOOS=linux GOARCH=amd64 $(GOENV) $(GO) build -ldflags="$(LDFLAGS)" -o $(OUT_DIR)/$(APP)-linux-amd64 ./cmd/$(APP)/
	@echo "  → $(OUT_DIR)/$(APP)-linux-amd64"

cross-arm64:
	@mkdir -p $(OUT_DIR)
	GOOS=linux GOARCH=arm64 $(GOENV) $(GO) build -ldflags="$(LDFLAGS)" -o $(OUT_DIR)/$(APP)-linux-arm64 ./cmd/$(APP)/
	@echo "  → $(OUT_DIR)/$(APP)-linux-arm64"

cross-mipsle:
	@mkdir -p $(OUT_DIR)
	GOOS=linux GOARCH=mipsle GOMIPS=softfloat $(GOENV) $(GO) build -ldflags="$(LDFLAGS)" -o $(OUT_DIR)/$(APP)-linux-mipsle ./cmd/$(APP)/
	@echo "  → $(OUT_DIR)/$(APP)-linux-mipsle"

cross-armv7:
	@mkdir -p $(OUT_DIR)
	GOOS=linux GOARCH=arm GOARM=7 $(GOENV) $(GO) build -ldflags="$(LDFLAGS)" -o $(OUT_DIR)/$(APP)-linux-armv7 ./cmd/$(APP)/
	@echo "  → $(OUT_DIR)/$(APP)-linux-armv7"

# Router-optimized builds (UPX compressed)
router-builds: cross
	@echo "=== Compressing router builds with UPX ==="
	@which upx > /dev/null && upx --best $(OUT_DIR)/$(APP)-linux-arm64 || echo "UPX not installed, skipping compression"
	@which upx > /dev/null && upx --best $(OUT_DIR)/$(APP)-linux-mipsle || echo "UPX not installed, skipping compression"
	@which upx > /dev/null && upx --best $(OUT_DIR)/$(APP)-linux-armv7 || echo "UPX not installed, skipping compression"
	@ls -lh $(OUT_DIR)/

# Show binary sizes
size:
	@ls -lh $(OUT_DIR)/
