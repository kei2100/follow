PACKAGES := $(shell go list ./...)

GREP := grep
ifeq ($(OS),Windows_NT)
	GREP := findstr
endif

# setup tasks
.PHONY: setup

setup:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u golang.org/x/lint/golint
	go get -u github.com/kyoh86/richgo
	go get -u github.com/mattn/go-colorable/cmd/colorable
	go mod tidy

# development tasks
.PHONY: fmt lint vet test test.nocache

fmt:
	goimports -w .

lint:
	goimports -d . | $(GREP) "^" && exit 1 || exit 0
	golint -set_exit_status $(PACKAGES)

vet:
	go vet $(PACKAGES)

test:
	richgo test -v -race $(PACKAGES) | colorable

test.nocache:
	richgo test -count=1 -v -race $(PACKAGES) | colorable
