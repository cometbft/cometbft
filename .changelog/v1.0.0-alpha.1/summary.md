*December 4, 2023*

This is a major release of CometBFT that includes several substantial changes
that aim to improve integrators' experience and increase the velocity of the
CometBFT development team, including:

1. The first officially supported release of the [data companion
   API](./docs/architecture/adr-101-data-companion-pull-api.md).
2. Versioning of both the Protobuf definitions _and_ RPC. By versioning our
   APIs, we aim to provide a level of commitment to API stability while
   simultaneously affording ourselves the ability to roll out substantial
   changes in non-breaking releases of CometBFT. See [ADR
   103](./docs/architecture/adr-103-proto-versioning.md) and [ADR
   107](./docs/architecture/adr-107-betaize-proto-versions.md).
3. Moving many Go packages that are currently publicly accessible into the
   `internal` directory such that the team can roll out substantial changes in
   future without needing to worry about causing breakages in users' codebases.
   The massive surface area of previous versions has in the past significantly
   hampered the team's ability to roll out impactful new changes to users, as
   previously such changes required a new breaking release (which currently
   takes 6 to 12 months to reach production use for many users). See [ADR
   109](./docs/architecture/adr-109-reduce-go-api-surface.md) for more details.

None of these changes are state machine-breaking for CometBFT-based networks,
but could be breaking for some users who depend on the Protobuf definitions type
URLs. See the [upgrading guidelines](./UPGRADING.md) and specific changes below
for more details.

**NB: This version is still an alpha-series release, which means that
API-breaking changes might still be introduced until such time that a _release
candidate_ is cut.** See [RELEASES.md](./RELEASES.md) for more information on
the stability guarantees we provide for pre-releases.
