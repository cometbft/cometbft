# CometBFT Tests

## Unit tests

The unit tests (ie. `go test`) can be run with `make test` from the root directory of the repository.

## Integration tests

The integration tests can be run with `make test_integrations` from the root directory of the repository.

Running the integrations test will build a docker container with local version of CometBFT
and run the following tests in docker containers:

- go tests, with --race
    - includes test coverage
- app tests
    - kvstore app over socket
- persistence tests
    - crash cometbft at each of many predefined points, restart, and ensure it syncs properly with the app

## End-to-end tests

You can run e2e nightly tests locally by running:

```sh
cd test/e2e
make && ./build/generator -g 5 -d networks/nightly/ -p && ./run-multiple.sh networks/nightly/*-group*-*.toml
```

If you just want a simple 4-node network, you can run:

```sh
cd test/e2e
make && ./build/runner -f networks/simple.toml
```

Please refer to the [README.MD](e2e/README.md) in the `e2e` folder for more information.

## Fuzzing

[Fuzzing](https://en.wikipedia.org/wiki/Fuzzing) of various system inputs.

See `./fuzz/README.md` for more details.
