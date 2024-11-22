#!/bin/sh
#
# Invoke Mockery v2 to update generated mocks for the given type.
# Last change was made based on changes for main in https://github.com/tendermint/tendermint/pull/9196 

go run github.com/vektra/mockery/v2@latest --disable-version-string --case underscore --name "$*"
<<<<<<< HEAD
=======

>>>>>>> cdab56d7d (feat(mocks): `mockery` v2.49.0 changes mock format (#4521))
