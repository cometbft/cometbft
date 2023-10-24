- `[abci]` Renamed the alias types for gRPC requests, responses, and service
  instances to follow the naming changes in the proto-derived
  `api/cometbft/abci/v1beta4` package
  ([\#1533](https://github.com/cometbft/cometbft/pull/1533)):
  * The prefixed naming pattern `RequestFoo`, `ReponseFoo` changed to
    suffixed `FooRequest`, `FooResponse`.
  * `ABCIClient` renamed to `ABCIServiceClient`.
  * `ABCIServer` renamed to `ABCIServiceServer`.
