#!/bin/sh
#
# Invoke Mockery v2 to update generated mocks for the given type.
# Last change was made based on changes for main in https://github.com/tendermint/tendermint/pull/9196 


<<<<<<< HEAD
go run github.com/vektra/mockery/v2 --disable-version-string --case underscore --name "$*"
=======


go run github.com/vektra/mockery/v2@latest --disable-version-string --case underscore --name "$*"

>>>>>>> 2bbb5c109 (misc(tools)!: remove `tools` package (#2046))
