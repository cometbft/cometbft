#!/bin/sh
#
# Invoke Mockery v2 to update generated mocks for the given type.
#
# The Go toolchain is pinned because newer Go releases change how the generated
# import blocks are formatted, which makes the check-mocks-metrics CI job fail.

GOTOOLCHAIN=go1.24.13 go run github.com/vektra/mockery/v2@v2.53.0 --name "$*"
