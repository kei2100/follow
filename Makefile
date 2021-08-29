GO ?= go

PACKAGES := $(shell $(GO) list ./...)

GREP := grep
ifeq ($(OS),Windows_NT)
	GREP := findstr
endif

# setup tasks
.PHONY: setup

setup:
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install golang.org/x/lint/golint@latest
	$(GO) install github.com/rakyll/gotest@latest
	$(GO) mod tidy

# development tasks
.PHONY: fmt lint vet test test.nocache

fmt:
	goimports -w .

lint:
	goimports -d . | $(GREP) "^" && exit 1 || exit 0
	golint -set_exit_status $(PACKAGES)

vet:
	$(GO) vet $(PACKAGES)

test:
	gotest -v -race $(PACKAGES)

test.nocache:
	gotest -count=1 -v -race $(PACKAGES)
