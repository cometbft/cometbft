mds = p2p.md mempool.md flood.md
quints = $(mds:.md=.qnt)

%.qnt: %.md
	lmt $<

all: $(quints)
.PHONY: all
 