#!/bin/sh
#
# Invoke Mockery v2 to update generated mocks for the given type.
# Last change was made based on changes for main in https://github.com/tendermint/tendermint/pull/9196 

<<<<<<< HEAD
go run github.com/vektra/mockery/v2@latest --disable-version-string --case underscore --name "$*"
=======
go run github.com/vektra/mockery/v2@v2.49.2 --disable-version-string --case underscore --name "$*"
>>>>>>> 13d852b43 (build(deps): Mock generation no longer uses the `latest` tag on `mockery`. (#4605))
