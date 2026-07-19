# Build the leitwerk gate. The gate is a single static Go binary; the checks it
# orchestrates stay as shell scripts under checks/ and are embedded into the
# binary (see assets.go) so it also runs with no sibling files on disk.
#
#   make build     -> bin/leitwerk (static, CGO disabled)
#   make test      -> unit + integration tests
#   make vet       -> go vet
#   make install   -> `go install` the binary into GOBIN
#   make clean
# Resolve `go`: prefer mise (the repo's pinned toolchain, per mise.toml) so a stale
# shim from another version manager on PATH is never used; fall back to a bare `go`
# where there is no mise (CI's setup-go). Override with `make GO=…` if needed.
GO ?= $(shell if command -v mise >/dev/null 2>&1; then echo "mise exec -- go"; else echo go; fi)
BIN := bin/leitwerk
PKG := ./cmd/leitwerk

.PHONY: build test vet install clean

build:
	CGO_ENABLED=0 $(GO) build -o $(BIN) $(PKG)

test:
	CGO_ENABLED=0 $(GO) test ./...

vet:
	$(GO) vet ./...

install:
	CGO_ENABLED=0 $(GO) install $(PKG)

clean:
	rm -f $(BIN)
