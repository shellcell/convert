GO ?= go
APP := convert
CMD := ./cmd/convert
DIST := dist
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || printf dev)
GOARM ?= 7
LDFLAGS := -s -w -X main.version=$(VERSION)

RELEASE_TARGETS := \
	linux_x86 \
	linux_x86_64 \
	linux_arm \
	linux_arm64 \
	macos_x86_64 \
	macos_arm64

.PHONY: build build-debug test vet clean release build-release checksums

build:
	$(GO) build -trimpath -ldflags="-s -w" -o $(APP) $(CMD)

build-debug:
	$(GO) build -o $(APP) $(CMD)

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

release: clean build-release checksums

build-release: $(RELEASE_TARGETS)

$(DIST):
	mkdir -p "$(DIST)"

define release_target
$(1): | $(DIST)
	@printf 'building %s\n' "$(APP)_$(VERSION)_$(2)"
	rm -rf "$(DIST)/$(APP)_$(VERSION)_$(2)"
	CGO_ENABLED=0 GOOS=$(3) GOARCH=$(4) $(if $(5),GOARM=$(5),) go build -trimpath -ldflags="$(LDFLAGS)" -o "$(DIST)/$(APP)_$(VERSION)_$(2)/$(APP)" $(CMD)
	cp LICENSE "$(DIST)/$(APP)_$(VERSION)_$(2)/LICENSE"
	tar -C "$(DIST)" -czf "$(DIST)/$(APP)_$(VERSION)_$(2).tar.gz" "$(APP)_$(VERSION)_$(2)"
endef

$(eval $(call release_target,linux_x86,linux_x86,linux,386,))
$(eval $(call release_target,linux_x86_64,linux_x86_64,linux,amd64,))
$(eval $(call release_target,linux_arm,linux_arm,linux,arm,$(GOARM)))
$(eval $(call release_target,linux_arm64,linux_arm64,linux,arm64,))
$(eval $(call release_target,macos_x86_64,macos_x86_64,darwin,amd64,))
$(eval $(call release_target,macos_arm64,macos_arm64,darwin,arm64,))

checksums: build-release
	@if command -v shasum >/dev/null 2>&1; then \
		cd "$(DIST)" && shasum -a 256 *.tar.gz > checksums.txt; \
	else \
		cd "$(DIST)" && sha256sum *.tar.gz > checksums.txt; \
	fi

clean:
	rm -f $(DIST)
