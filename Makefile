.DEFAULT_GOAL = help

vendor      := vendor
target      := target
bin         := $(target)/bin
reports     := $(target)/reports

## build: Compile binaries.
go_src := $(shell find * -name *.go -not -path "$(vendor)/*" -not -path "$(target)/*")
go_out := $(patsubst cmd/%/main.go,$(bin)/%,$(wildcard cmd/*/main.go))

.PHONY: build
build: $(go_out)

$(bin)/%: cmd/%/main.go $(go_src) | $(bin)
	@go build --trimpath --ldflags='-X "main.version=$(version)"' -o=$@ $<

$(bin):
	@mkdir -p $@

$(reports):
	@mkdir -p $@

## install: Installs all packages.
.PHONY: install
install:
	@go install ./...

## generate: Run generators.
.PHONY: generate
generate: go/generate

.PHONY: go/generate
go/generate:
	@go generate ./...

## lint: Run static analysis.
.PHONY: lint
lint: go/lint

.PHONY: go/lint
go/lint:
	@golangci-lint run

## tests: Run tests.
.PHONY: test tests
test tests: go/test go/race

.PHONY: go/race
go/race: $(go_src)
	@go test -short -race -count=100 ./...

.PHONY: go/test
go/test: $(go_src) | $(reports)
	@go test -v -covermode=atomic -coverprofile=$(reports)/cover.out ./...

## clean: Remove created resources.
.PHONY: clean
clean:
	@rm -rf $(vendor) $(target)

## version: Display current version.
git_tag := $(shell git describe --tags 2>/dev/null)
git_sha := $(shell git rev-parse HEAD 2>/dev/null)
version := $(if $(git_tag),$(git_tag),(unknown on $(git_sha)))

.PHONY: version
version:
	@echo "version $(version)"

## help: Display available targets.
.PHONY: help
help: Makefile
	@echo "Usage: make [target]"
	@echo
	@echo "Targets:"
	@sed -n 's/^## //p' $< | awk -F ':' '{printf "  %-20s%s\n",$$1,$$2}'
