all:
	@echo "Targets:"
	@echo "\ttraces"
	@echo "\ttypecheck"
	@echo "\ttlc"
	@echo "\tclean"

TLA_MODULE := MempoolV0MC
TLC_JAR := ~/dev/tla/tools/tla2tools.jar

.PHONY: traces
traces:
	time ./generate-traces.sh

.PHONY: typecheck
typecheck:
	apalache-mc typecheck --features=no-rows $(TLA_MODULE).tla

.PHONY: tlc
tlc:
	java -XX:+UseParallelGC -classpath $(TLC_JAR) tlc2.TLC $(TLA_MODULE)

.PHONY: clean
clean:
	rm -f ./*_TTrace_*.tla
	rm -frd ./states
	rm -frd ./_apalache-out
