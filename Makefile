# Binary name and entry point
BINARY_NAME=httping-ng
MAIN_PKG=./cmd/httping

# Build dirs
BUILD_DIR=build
DIST_DIR=dist

# OS/ARCH targets
OS_LIST=linux darwin windows
ARCH_LIST=amd64 arm64

# Default native build
all:
	go build -o $(BINARY_NAME) $(MAIN_PKG)

# Cross-compile all OS/ARCH combinations
build-all:
	@mkdir -p $(BUILD_DIR)
	@for GOOS in $(OS_LIST); do \
		for GOARCH in $(ARCH_LIST); do \
			EXT=$${GOOS=="windows" && echo ".exe" || echo ""}; \
			OUT=$(BUILD_DIR)/$(BINARY_NAME)-$${GOOS}-$${GOARCH}$$EXT; \
			echo "Building $$OUT..."; \
			GOOS=$$GOOS GOARCH=$$GOARCH CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $$OUT $(MAIN_PKG); \
		done \
	done

# Release: archive all binaries into zips/tars
release: clean build-all
	@mkdir -p $(DIST_DIR)
	@cd $(BUILD_DIR) && \
	for f in $(BINARY_NAME)-*; do \
		base=$$(basename $$f); \
		if echo $$base | grep -q 'windows'; then \
			zip -j ../$(DIST_DIR)/$$base.zip $$base; \
		else \
			tar -czf ../$(DIST_DIR)/$$base.tar.gz -C . $$base; \
		fi; \
	done

# Clean binaries, builds, and archives
clean:
	rm -rf $(BINARY_NAME) $(BUILD_DIR) $(DIST_DIR)

.PHONY: all build-all release clean
