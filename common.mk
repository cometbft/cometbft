# This contains Makefile logic that is common to several makefiles

BUILD_TAGS ?= cometbft

COMMIT_HASH := $(shell git rev-parse --short HEAD)
LD_FLAGS = -X github.com/cometbft/cometbft/version.TMGitCommitHash=$(COMMIT_HASH)
BUILD_FLAGS = -mod=readonly -ldflags "$(LD_FLAGS)"
# allow users to pass additional flags via the conventional LDFLAGS variable
LD_FLAGS += $(LDFLAGS)

# handle nostrip
ifeq (,$(findstring nostrip,$(COMETBFT_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
  LD_FLAGS += -s -w
endif

# handle race
ifeq (race,$(findstring race,$(COMETBFT_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_FLAGS += -race
endif

# handle cleveldb
ifeq (cleveldb,$(findstring cleveldb,$(COMETBFT_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_TAGS += cleveldb
endif

# handle badgerdb
ifeq (badgerdb,$(findstring badgerdb,$(COMETBFT_BUILD_OPTIONS)))
  BUILD_TAGS += badgerdb
endif

# handle rocksdb
ifeq (rocksdb,$(findstring rocksdb,$(COMETBFT_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_TAGS += rocksdb
endif

# handle boltdb
ifeq (boltdb,$(findstring boltdb,$(COMETBFT_BUILD_OPTIONS)))
  BUILD_TAGS += boltdb
endif
