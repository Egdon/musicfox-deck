PACKAGE_ROOT          := $(CURDIR)
INJECT_PACKAGE        ?= github.com/go-musicfox/go-musicfox/internal/types
LDFLAGS               := -s -w
LASTFM_KEY            ?=
LASTFM_SECRET         ?=

# modvendor 用于在 go mod vendor 后复制 C/C++ 头文件等非 Go 文件
MODVENDOR_BIN := $(shell go env GOPATH)/bin/modvendor

# ── OS 检测 ──────────────────────────────────────────────────────────────────
# Windows_NT → 使用 PowerShell 构建脚本；其他系统 → 使用 Bash 构建脚本
ifeq ($(OS),Windows_NT)
    UNAME_S := Windows_NT
else
    UNAME_S := $(shell uname -s)
endif
IS_MACOS := $(filter Darwin,$(UNAME_S))

ifeq ($(OS),Windows_NT)
    BUILD_SCRIPT := powershell -NoProfile -ExecutionPolicy Bypass -File hack\build.ps1
    NULL_DEV     := nul
    WHICH_CMD    := where
else
    BUILD_SCRIPT := $(PACKAGE_ROOT)/hack/build.sh
    NULL_DEV     := /dev/null
    WHICH_CMD    := which
endif

.PHONY: build
build:
ifeq ($(OS),Windows_NT)
	$(BUILD_SCRIPT) build -BuildTags "enable_global_hotkey,purego"
else
	BUILD_TAGS="enable_global_hotkey,purego" $(BUILD_SCRIPT) build
endif

# build-macapp 仅在 macOS 下有效
.PHONY: build-macapp
build-macapp:
ifeq ($(IS_MACOS),Darwin)
	@mkdir -p $(PACKAGE_ROOT)/bin
	BUILD_TAGS="enable_global_hotkey,purego" $(BUILD_SCRIPT) build
	@mkdir -p $(PACKAGE_ROOT)/bin/tapefox.app/Contents/MacOS
	@mkdir -p $(PACKAGE_ROOT)/bin/tapefox.app/Contents/Resources
	@cp $(PACKAGE_ROOT)/bin/tapefox $(PACKAGE_ROOT)/bin/tapefox.app/Contents/MacOS/go-musicfox
	@cp $(PACKAGE_ROOT)/deploy/musicfox-app.wrapper.sh $(PACKAGE_ROOT)/bin/tapefox.app/Contents/MacOS/tapefox
	@cp $(PACKAGE_ROOT)/deploy/musicfox.app/Contents/Info.plist $(PACKAGE_ROOT)/bin/tapefox.app/Contents/Info.plist
	@cp $(PACKAGE_ROOT)/deploy/musicfox.app/Contents/Resources/Musicfox.icns $(PACKAGE_ROOT)/bin/tapefox.app/Contents/Resources/Musicfox.icns
	@chmod +x $(PACKAGE_ROOT)/bin/tapefox.app/Contents/MacOS/go-musicfox
	@chmod +x $(PACKAGE_ROOT)/bin/tapefox.app/Contents/MacOS/tapefox
	@xattr -r -d com.apple.quarantine -- $(PACKAGE_ROOT)/bin/tapefox.app
else
	@echo "build-macapp is only supported on macOS"
endif

.PHONY: init
init:
	git config --local core.hooksPath githooks

.PHONY: install
install:
ifeq ($(OS),Windows_NT)
	$(BUILD_SCRIPT) install -BuildTags "enable_global_hotkey,purego"
else
	BUILD_TAGS="enable_global_hotkey,purego" $(BUILD_SCRIPT) install
endif

.PHONY: scoop-config-gen
scoop-config-gen:
ifneq ($(OS),Windows_NT)
	$(PACKAGE_ROOT)/hack/scoop_gen.sh
else
	@echo "Please run 'hack/scoop_gen.sh' in Git Bash or WSL on Windows"
endif

.PHONY: changelog-gen
changelog-gen:
ifneq ($(OS),Windows_NT)
	$(PACKAGE_ROOT)/hack/changelog_gen.sh
else
	@echo "Please run 'hack/changelog_gen.sh' in Git Bash or WSL on Windows"
endif

.PHONY: lint
lint:
	golangci-lint run -v

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix -v

.PHONY: test
test:
	go test ./internal/... ./utils/... \
		-coverpkg=./internal/...,./utils/... \
		-covermode=atomic -coverprofile=coverage.txt

.PHONY: clean
clean:
ifeq ($(OS),Windows_NT)
	@powershell -NoProfile -Command "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue '$(PACKAGE_ROOT)/bin', '$(PACKAGE_ROOT)/vendor', '$(PACKAGE_ROOT)/coverage.txt'"
else
	@rm -rf $(PACKAGE_ROOT)/bin $(PACKAGE_ROOT)/vendor $(PACKAGE_ROOT)/coverage.txt
endif

.PHONY: modvendor-tool
modvendor-tool:
	@$(WHICH_CMD) modvendor >$(NULL_DEV) 2>&1 || ( \
		echo "Installing modvendor..."; \
		go install github.com/goware/modvendor@latest; \
	)

.PHONY: vendor
vendor: modvendor-tool
	go mod tidy
	go mod vendor
	@echo "Copying C/C++ header files for CGo dependencies..."
	modvendor -copy="**/*.h **/*.c" -v

# release-snapshot 在 Ubuntu 22.04 容器里试做 Linux 与 Windows 的发布包，输出到 dist/；
# 正式发布由 .github/workflows/release.yaml 完成
CONTAINER_ENGINE ?= $(shell command -v podman 2>/dev/null || command -v docker 2>/dev/null)

.PHONY: release-snapshot
release-snapshot:
ifneq ($(OS),Windows_NT)
	$(CONTAINER_ENGINE) run --rm --network host --security-opt label=disable \
		-e HOST_UID=$$(id -u) -e HOST_GID=$$(id -g) \
		-v $(PACKAGE_ROOT):/src -w /src docker.io/library/ubuntu:22.04 \
		hack/release_container.sh --snapshot
else
	@echo "release-snapshot is not supported on Windows"
endif
