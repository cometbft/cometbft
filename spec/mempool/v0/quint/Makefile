all:
	@echo "Targets:"
	@echo "\ttypecheck"
	@echo "\tparse"
	@echo "\ttest"
	@echo "\tverify"
	@echo "\ttraces"
	@echo "\tclean"

TRACES_DIR := traces

.PHONY: typecheck
typecheck:
	quint typecheck tests.qnt

parse:
	quint parse tests.qnt
.PHONY: parse

test:
	mkdir -p ${TRACES_DIR}/tests
	quint test --verbosity=3 --main tests --match test --max-samples 3 --output "${TRACES_DIR}/tests/{}_trace{#}.itf.json" tests.qnt
.PHONY: test

verify:
	mkdir -p ${TRACES_DIR}/verify
	quint verify --out-itf "${TRACES_DIR}/verify/{}_trace{#}.itf.json" tests.qnt
.PHONY: verify

traces:
	time ./gen-traces.sh
.PHONY: traces

clean:
	rm -frd ./traces
.PHONY: clean
