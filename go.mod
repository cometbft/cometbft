module github.com/tendermint/tendermint

go 1.22.0
toolchain go1.22.9

require (
	github.com/BurntSushi/toml v1.4.0
	github.com/ChainSafe/go-schnorrkel v1.1.0
	github.com/Workiva/go-datastructures v1.1.5
	github.com/adlio/schema v1.3.6
	github.com/fortytw2/leaktest v1.3.0
	github.com/go-kit/kit v0.13.0
	github.com/go-kit/log v0.2.1
	github.com/go-logfmt/logfmt v0.6.0
	github.com/gofrs/uuid v4.4.0+incompatible
	github.com/google/orderedcode v0.0.1
	github.com/gorilla/websocket v1.5.3
	github.com/gtank/merlin v0.1.1
	github.com/lib/pq v1.10.9
	github.com/libp2p/go-buffer-pool v0.1.0
	github.com/minio/highwayhash v1.0.3
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.20.5
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
	github.com/rs/cors v1.11.1
	github.com/sasha-s/go-deadlock v0.3.5
	github.com/snikch/goodman v0.0.0-20171125024755-10e37e294daa
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.19.0
	github.com/stretchr/testify v1.10.0
	golang.org/x/crypto v0.28.0
	golang.org/x/net v0.30.0
	google.golang.org/grpc v1.68.0
)

require github.com/google/uuid v1.6.0

require github.com/gogo/protobuf v1.3.2

require (
	github.com/creachadair/taskgroup v0.13.0
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
)

require (
	github.com/Masterminds/semver/v3 v3.3.1
	github.com/btcsuite/btcd/btcec/v2 v2.3.4
	github.com/btcsuite/btcd/btcutil v1.1.6
	github.com/cometbft/cometbft-db v0.9.5
	github.com/cometbft/cometbft-load-test v0.2.0
	github.com/go-git/go-git/v5 v5.12.0
	github.com/golang/protobuf v1.5.4
	github.com/ory/dockertest v3.3.5+incompatible
	golang.org/x/sync v0.8.0
	gonum.org/v1/gonum v0.15.0
	google.golang.org/protobuf v1.35.1
)

// TODO: Remove once these libs are updated.
require github.com/opencontainers/runc v1.2.0-rc.2 // indirect

replace (
	github.com/adlio/schema v1.3.6 => github.com/informalsystems/schema v0.0.0-20240910074420-6f3a29ec8110

	github.com/go-kit/kit v0.13.0 => github.com/informalsystems/kit v0.0.0-20240910100206-37b6b26dc6fd

	github.com/spf13/afero v1.11.0 => github.com/informalsystems/afero v0.0.0-20240822164725-478adf757a7e

	gonum.org/v1/gonum => gonum.org/v1/gonum v0.15.2-0.20240904063137-1ca563a018b6
)

require (
	dario.cat/mergo v1.0.1 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/ProtonMail/go-crypto v1.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudflare/circl v1.3.9 // indirect
	github.com/containerd/continuity v0.4.3 // indirect
	github.com/cosmos/go-bip39 v0.0.0-20180819234021-555e2067c45d // indirect
	github.com/cyphar/filepath-securejoin v0.2.5 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/dgraph-io/badger/v2 v2.2007.4 // indirect
	github.com/dgraph-io/ristretto v0.1.1 // indirect
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/docker/cli v27.0.3+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.5.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.1.0 // indirect
	github.com/golang/glog v1.2.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/gotestyourself/gotestyourself v1.4.0 // indirect
	github.com/gtank/ristretto255 v0.1.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jmhodges/levigo v1.0.0 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/linxGnu/grocksdb v1.9.3 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mimoo/StrobeGo v0.0.0-20220103164710-9a04d6ca976b // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/sys/user v0.1.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/petermattis/goid v0.0.0-20240813172612-4fcff4a6cae7 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.59.1 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/sagikazarmark/locafero v0.6.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/skeema/knownhosts v1.2.2 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	go.etcd.io/bbolt v1.3.11 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20240909161429-701f63a606c0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
