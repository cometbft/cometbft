#!/bin/sh
#
# Invoke Mockery v2 to update generated mocks for the given type.

go run github.com/vektra/mockery/v2@v2.53.5 --disable-version-string --case underscore --all
