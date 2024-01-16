#!/bin/sh
#
# Invoke Mockery v2 to update generated mocks for the given type.



go install github.com/vektra/mockery/v2@latest
mockery --disable-version-string --case underscore --name "$*"

