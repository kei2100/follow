PACKAGES := $(shell go list ./...)

GREP := grep
ifeq ($(OS),Windows_NT)
	GREP := findstr
endif

# setup tasks
.PHONY: setup

setup:
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/lint/golint@latest
	go install github.com/rakyll/gotest@latest
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
	gotest -v -race $(PACKAGES)

test.nocache:
	gotest -count=1 -v -race $(PACKAGES)
