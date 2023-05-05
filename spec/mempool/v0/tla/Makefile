all:
	@echo "Targets:"
	@echo "\ttraces"
	@echo "\ttypecheck"
	@echo "\tclean"

.PHONY: traces
traces:
	time ./generate-traces.sh

.PHONY: typecheck
typecheck:
	apalache-mc typecheck --features=no-rows MempoolV0MC.tla

.PHONY: clean
clean:
	rm -f ./*_TTrace_*.tla
	rm -frd ./states
	rm -frd ./_apalache-out
