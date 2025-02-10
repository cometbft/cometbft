# This contains Makefile logic that is common to several makefiles

BUILD_TAGS ?= cometbft

COMMIT_HASH := $(shell git rev-parse --short HEAD)
LD_FLAGS = -X github.com/cometbft/cometbft/version.CMTGitCommitHash=$(COMMIT_HASH)
BUILD_FLAGS = -mod=readonly -ldflags "$(LD_FLAGS)"
# allow users to pass additional flags via the conventional LDFLAGS variable
LD_FLAGS += $(LDFLAGS)

# handle nostrip
ifeq (nostrip,$(findstring nostrip,$(COMETBFT_BUILD_OPTIONS)))
  #prepare for delve
  BUILD_FLAGS+= -gcflags "all=-N -l"
else
  BUILD_FLAGS += -trimpath
  LD_FLAGS += -s -w
endif

# handle race
ifeq (race,$(findstring race,$(COMETBFT_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_FLAGS += -race
endif

# handle clock_skew
ifeq (clock_skew,$(findstring clock_skew,$(COMETBFT_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_TAGS += clock_skew
endif

# handle bls12381
ifeq (bls12381,$(findstring bls12381,$(COMETBFT_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_TAGS += bls12381
endif

# handle secp256k1eth
ifeq (secp256k1eth,$(findstring secp256k1eth,$(COMETBFT_BUILD_OPTIONS)))
  BUILD_TAGS += secp256k1eth
endif

# handle nodebug
ifeq (nodebug,$(findstring nodebug,$(COMETBFT_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_TAGS += nodebug
endif
