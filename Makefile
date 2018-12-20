ifeq ($(origin VERSION), undefined)
	VERSION = `git rev-parse --short HEAD`
endif

HOST_GOOS   ?= $(shell go env GOOS)
HOST_GOARCH ?= $(shell go env GOARCH)
REPOPATH    ?= github.com/alternative-storage/torus

export GOPATH := $(realpath ../../../..)
#                              ^  ^  ^~ GOPATH
#                              |  |~ GOPATH/src
#                              |~ GOPATH/src/github.com

VERBOSE_1 := -v
VERBOSE_2 := -v -x
WHAT      := torusd torusctl torusblk mkfs.torus fsck.torus
GLIDE     := ./tools/glide
GLIDE_V   := v0.12.3

.PHONY: build
build: vendor
	for target in $(WHAT); do \
		echo "building $$target..."; \
		$(BUILD_ENV_FLAGS) go build $(VERBOSE_$(V)) -o bin/$$target -ldflags "-X $(REPOPATH).Version=$(VERSION)" ./cmd/$$target; \
	done

.PHONY: test
test: tools/glide
	go test --race $(shell $(GLIDE) novendor)

.PHONY: bench
bench: tools/glide
	go install ./vendor/github.com/cespare/prettybench
	go test -bench $(shell $(GLIDE) novendor) | $(GOPATH)/bin/prettybench

.PHONY: vet
vet: tools/glide
	go vet $(shell $(GLIDE) novendor)

.PHONY: fmt
fmt: tools/glide
	go fmt $(shell $(GLIDE) novendor)

.PHONY: lint
lint:
	@for dir in $(shell $(GLIDE) novendor); do \
		golint $$dir; \
	done;

.PHONY: clean
clean:
	rm -rf ./local-cluster ./bin/torus*

.PHONY: cleanall
cleanall: clean
	rm -rf /tmp/etcd bin tools vendor

.PHONY: run3
run3:
	goreman start

.PHONY: relase
release: releasetar
	goxc -d ./release -tasks-=go-vet,go-test -os="linux darwin" -pv=$(VERSION)  -arch="386 amd64 arm arm64" -build-ldflags="-X $(REPOPATH).Version=$(VERSION)" -resources-include="README.md,Documentation,LICENSE,contrib" -main-dirs-exclude="vendor,cmd/ringtool"

.PHONY: releasetar
releasetar:
	mkdir -p release/$(VERSION)
	glide install --strip-vcs --strip-vendor --update-vendored --delete
	glide-vc --only-code --no-tests --keep="**/*.json.in"
	git ls-files > /tmp/torusbuild
	find vendor >> /tmp/torusbuild
	tar -cvf release/$(VERSION)/torus_$(VERSION)_src.tar -T /tmp/torusbuild --transform 's,^,torus_$(VERSION)/,'
	rm /tmp/torusbuild
	gzip release/$(VERSION)/torus_$(VERSION)_src.tar

vendor: tools/glide
	$(GLIDE) install

tools/glide:
	@echo "Downloading glide"
	mkdir -p tools
	curl -L https://github.com/Masterminds/glide/releases/download/$(GLIDE_V)/glide-$(GLIDE_V)-$(HOST_GOOS)-$(HOST_GOARCH).tar.gz | tar -xz -C tools
	mv tools/$(HOST_GOOS)-$(HOST_GOARCH)/glide $(GLIDE)
	rm -r tools/$(HOST_GOOS)-$(HOST_GOARCH)

help:
	@echo "Influential make variables"
	@echo "  V                 - Build verbosity {0,1,2}."
	@echo "  BUILD_ENV_FLAGS   - Environment added to 'go build'."
	@echo "  WHAT              - Command to build. (e.g. WHAT=torusctl)"
