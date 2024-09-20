#!/usr/bin/make -f

########################################
### Testing

BINDIR ?= $(GOPATH)/bin

## required to be run first by most tests
build_docker_test_image:
	docker build -t tester -f ./test/docker/Dockerfile .
.PHONY: build_docker_test_image

### coverage, app, persistence, and libs tests
test_cover:
	# run the go unit tests with coverage
	bash test/test_cover.sh
.PHONY: test_cover

test_apps:
	# run the app tests using bash
	# requires `abci-cli` and `cometbft` binaries installed
	bash test/app/test.sh
.PHONY: test_apps

test_abci_apps:
	bash abci/tests/test_app/test.sh
.PHONY: test_abci_apps

test_abci_cli:
	# test the cli against the examples in the tutorial at:
	# ./docs/abci-cli.md
	# if test fails, update the docs ^
	@ bash abci/tests/test_cli/test.sh
.PHONY: test_abci_cli

test_integrations:
	make build_docker_test_image
	make install
	make install_abci
	make test_cover
	make test_apps
	make test_abci_apps
	make test_abci_cli
.PHONY: test_integrations

test_release:
	@go test -tags release $(PACKAGES)
.PHONY: test_release

### go tests
test:
	@echo "--> Running go test"
	@go test -p 1 $(PACKAGES) -tags bls12381
.PHONY: test

test_race:
	@echo "--> Running go test --race"
	@go test -p 1 -race $(PACKAGES) -tags bls12381
.PHONY: test_race

test_deadlock:
	@echo "--> Running go test with deadlock support"
	@go test -p 1 $(PACKAGES) -tags deadlock,bls12381
.PHONY: test_deadlock

# Implements test splitting and running. This is pulled directly from
# the github action workflows for better local reproducibility.

GO_TEST_FILES := $(shell find $(CURDIR) -name "*_test.go")

# default to four splits by default
NUM_SPLIT ?= 4

# The format statement filters out all packages that don't have tests.
# Note we need to check for both in-package tests (.TestGoFiles) and
# out-of-package tests (.XTestGoFiles).
$(BUILDDIR)/packages.txt:$(GO_TEST_FILES) $(BUILDDIR)
	go list -f "{{ if (or .TestGoFiles .XTestGoFiles) }}{{ .ImportPath }}{{ end }}" ./... | sort > $@

split-test-packages:$(BUILDDIR)/packages.txt
	split -d -n l/$(NUM_SPLIT) $< $<.

# Used by the GitHub CI, in order to run tests in parallel
test-group-%:split-test-packages
	cat $(BUILDDIR)/packages.txt.$*
	cat $(BUILDDIR)/packages.txt.$* | xargs go test -tags bls12381 -mod=readonly -timeout=400s -race -coverprofile=$(BUILDDIR)/$*.profile.out
