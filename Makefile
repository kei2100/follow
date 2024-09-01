PACKAGES := $(shell go list ./...)

GREP := grep
ifeq ($(OS),Windows_NT)
	GREP := findstr
endif

# setup tasks
.PHONY: setup

setup:
	go install golang.org/x/tools/cmd/goimports@latest

# development tasks
.PHONY: fmt lint test test.nocache

fmt:
	goimports -w .

lint:
	goimports -d . | $(GREP) "^" && exit 1 || exit 0
	go vet $(PACKAGES)

test:
	go test -race $(PACKAGES)

test.nocache:
	go test -count=1 -race $(PACKAGES)
