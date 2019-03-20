GOTOOLS := \
	github.com/alecthomas/gometalinter \
	github.com/git-chglog/git-chglog/cmd/git-chglog \
	github.com/golang/dep/cmd/dep \
	golang.org/x/tools/cmd/cover \
	github.com/stretchr/testify/assert \

DIRS     ?= $(shell find . -name '*.go' | grep --invert-match 'vendor' | xargs -n 1 dirname | sort --unique)
PKG_NAME ?= scenery

VERSION := $(shell git describe --tags --always --dirty="-dev")

LDFLAGS := -ldflags='-w -s -X "main.Version=$(VERSION)"'

COVERAGE_PROFILE ?= coverage.out

.DEFAULT_GOAL := help

.PHONY: build
build: ## Builds a local Go binary
	@echo "---> Building"
	CGO_ENABLED=0 go build -o ./bin/$(PKG_NAME) $(LDFLAGS)

.PHONY: clean
clean: ## Removes Go temporary build files build directory
	@echo "---> Cleaning"
	@rm -rf ./bin

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: html
html: ## Generates an HTML coverage report
	@echo "---> Generating HTML coverage report"
	go tool cover -html $(COVERAGE_PROFILE)

.PHONY: install_tools
install_tools: ## Installs all development tool dependencies
	@echo "--> Installing tools"
	go get -u $(GOTOOLS)
	gometalinter --install

.PHONY: lint
lint: ## Runs all linters
	@echo "---> Linting... this might take a minute"
	gometalinter --vendor --tests --deadline=3m $(LFLAGS) $(DIRS)

.PHONY: release
release: ## Creates a new release with the given tag
	@echo "---> Creating new release"
ifndef tag
	$(error tag must be specified)
endif
	git-chglog --output CHANGELOG.md --next-tag $(tag)
	git add CHANGELOG.md
	git commit -m $(tag)
	git tag $(tag)
	git push origin master --tags

.PHONY: test
test: ## Runs all the tests and outputs the coverage report
	@echo "---> Testing"
	go test ./... -coverprofile $(COVERAGE_PROFILE) $(TFLAGS)

.PHONY: uninstall_tools ## test
uninstall_tools: ## Uninstalls all development tool dependencies
	@echo "--> Uninstalling tools"
	go clean -i $(GOTOOLS)
