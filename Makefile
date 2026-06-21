GO ?= go
FMT_DIRS := prove examples rollup

.PHONY: verify build fmt vet test secrets tidy-check clean-keys

## verify: the full CI gate — formatting, vet, tests, and a secrets scan.
verify: fmt vet test secrets
	@echo "verify: OK"

## build: compile all active packages.
build:
	$(GO) build ./...

## fmt: fail if any tracked Go source is not gofmt-clean.
fmt:
	@unformatted=$$(gofmt -l $(FMT_DIRS)); \
	if [ -n "$$unformatted" ]; then \
		echo "gofmt needed on:"; echo "$$unformatted"; exit 1; \
	fi
	@echo "fmt: OK"

## vet: go vet across the module (nested legacy module is excluded).
vet:
	$(GO) vet ./...

## test: run the full test suite (real Groth16 prove/verify included).
test:
	$(GO) test ./...

## secrets: scan source and docs for committed secret material.
secrets:
	@./scripts/secrets-scan.sh

## tidy-check: ensure go.mod/go.sum are tidy.
tidy-check:
	$(GO) mod tidy
	@git diff --exit-code go.mod go.sum
