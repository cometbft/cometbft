mds = p2p.md mempool.md flood.md dog.md
quints = $(mds:.md=.qnt)

%.qnt: %.md
	lmt $<

all: $(quints)
.PHONY: all
 