#!/usr/bin/make -f

########################################
### Testing

BINDIR ?= $(GOPATH)/bin

#?test_apps: Run the app tests
test_apps: install install_abci
	@bash test/app/test.sh
.PHONY: test_apps

#?test_abci_cli: Test the cli against the examples in the tutorial at: ./docs/abci-cli.md
# if test fails, update the docs ^
test_abci_cli: install_abci
	@bash abci/tests/test_cli/test.sh
.PHONY: test_abci_cli

#?test_integrations: Runs all integration tests
test_integrations: test_apps test_abci_cli test_integrations_cleanup
.PHONY: test_integrations

#?test_integrations_cleanup: Cleans up the test data created by test_integrations
test_integrations_cleanup:
	@bash test/app/clean.sh
.PHONY: test_integrations_cleanup

test_release:
	@go test -tags release $(PACKAGES)
.PHONY: test_release

### go tests
test:
	@echo "--> Running go test"
	@go test -p 1 $(PACKAGES) -tags bls12381,secp256k1eth
.PHONY: test

test_race:
	@echo "--> Running go test --race"
	@go test -p 1 -race $(PACKAGES) -tags bls12381,secp256k1eth
.PHONY: test_race

test_deadlock:
	@echo "--> Running go test with deadlock support"
	@go test -p 1 $(PACKAGES) -tags deadlock,bls12381,secp256k1eth
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
	cat $(BUILDDIR)/packages.txt.$* | xargs go test -tags 'bls12381,secp256k1eth' -mod=readonly -timeout=400s -race -coverprofile=$(BUILDDIR)/$*.profile.out
