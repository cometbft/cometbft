package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type GasRow struct {
	Function string `json:"function"`
	Min      uint64 `json:"min"`
	Avg      uint64 `json:"avg"`
	Median   uint64 `json:"median"`
	Max      uint64 `json:"max"`
	Calls    uint64 `json:"calls"`
}

type CalldataRow struct {
	Name              string `json:"name"`
	ByteLength        uint64 `json:"byte_length"`
	ZeroBytes         uint64 `json:"zero_bytes"`
	NonzeroBytes      uint64 `json:"nonzero_bytes"`
	GasL1Standard     uint64 `json:"gas_l1_standard"`
	GasL1FloorEIP7623 uint64 `json:"gas_l1_floor_eip_7623"`
	BlobBytes         uint64 `json:"blob_bytes"`
}

type Benchmark struct {
	ID             string   `json:"id"`
	Group          string   `json:"group"`
	Question       string   `json:"question"`
	Model          string   `json:"model"`
	Assumptions    []string `json:"assumptions"`
	Fixture        string   `json:"fixture"`
	ExpectedOutput string   `json:"expected_output"`
	ExecutionRow   string   `json:"execution_row,omitempty"`
	CalldataRow    string   `json:"calldata_row,omitempty"`
}

type BlsDRow struct {
	Signers       uint64 `json:"signers"`
	Pairings      uint64 `json:"pairings"`
	MeasuredGas   uint64 `json:"measured_gas"`
	FormulaGas    uint64 `json:"formula_gas"`
	OverheadGas   int64  `json:"overhead_gas"`
	Formula       string `json:"formula"`
	Source        string `json:"source"`
	FoundryStable bool   `json:"foundry_stable"`
}

type Manifest struct {
	GeneratedAt string `json:"generated_at"`
	Toolchain   struct {
		Forge               string   `json:"forge"`
		Solc                []string `json:"solc"`
		Go                  string   `json:"go"`
		GitCommit           string   `json:"git_commit"`
		GitStatus           string   `json:"git_status"`
		EVMVersion          string   `json:"evm_version"`
		Optimizer           bool     `json:"optimizer"`
		OptimizerRuns       uint64   `json:"optimizer_runs"`
		ViaIR               bool     `json:"via_ir"`
		AutoDetectSolc      bool     `json:"auto_detect_solc"`
		CompilerPinningNote string   `json:"compiler_pinning_note"`
	} `json:"toolchain"`
	Commands         []string          `json:"commands"`
	FixtureHashes    map[string]string `json:"fixture_hashes"`
	Benchmarks       []Benchmark       `json:"benchmarks"`
	ExecutionGas     []GasRow          `json:"execution_gas"`
	Calldata         []CalldataRow     `json:"calldata"`
	BlsDMultiMessage []BlsDRow         `json:"bls_d_multi_message"`
	Ics23HappyPath   []GasRow          `json:"ics23_happy_path_gas"`
}

func main() {
	m := Manifest{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Commands: []string{
			"go run -tags bls12381 ./script",
			"scripts/update-report-data.sh",
			"forge test --summary",
			"forge test --gas-report --summary",
			"forge test --match-test testCalldataGasEstimates -vvvv",
			"forge fmt --check",
		},
		FixtureHashes: map[string]string{},
		Benchmarks:    benchmarkManifest(),
	}

	m.Toolchain.Forge = commandOutput("forge", "--version")
	m.Toolchain.Go = commandOutput("go", "version")
	m.Toolchain.GitCommit = commandOutput("git", "rev-parse", "HEAD")
	m.Toolchain.GitStatus = commandOutput("git", "status", "--short")
	m.Toolchain.CompilerPinningNote = "Source files use exact pragmas =0.8.30 and =0.6.12; foundry.toml uses auto_detect_solc=true because the benchmark intentionally compiles both pinned versions."

	config := readFile("generated/forge-config.json")
	if config == "" {
		config = commandOutput("forge", "config", "--json")
		_ = os.WriteFile("generated/forge-config.json", []byte(config), 0o644)
	}
	m.Toolchain.EVMVersion = jsonString(config, "evm_version")
	m.Toolchain.Optimizer = jsonBool(config, "optimizer")
	m.Toolchain.OptimizerRuns = jsonUint(config, "optimizer_runs")
	m.Toolchain.ViaIR = jsonBool(config, "via_ir")
	m.Toolchain.AutoDetectSolc = jsonBool(config, "auto_detect_solc")
	m.Toolchain.Solc = parseSolcVersions(readFile("generated/forge-build.txt"))

	m.ExecutionGas = parseGasRows(readFile("generated/gas-report.txt"))
	m.Calldata = parseCalldataRows(readFile("generated/calldata-report.txt"))
	m.BlsDMultiMessage = parseBlsD()
	m.Ics23HappyPath = parseIcs23HappyPath()
	hashFixtures(m.FixtureHashes)

	writeJSON("benchmarks.json", m)
	writeTables("generated/benchmark-tables.md", m)
}

func benchmarkManifest() []Benchmark {
	return []Benchmark{
		{
			ID: "compact-commit-50", Group: "secp256k1eth", ExecutionRow: "verifyCommitCompact", CalldataRow: "commit-compact-50", Fixture: "test/fixtures/update_50_equal.json",
			Question:       "What does the production-shape compact secp256k1eth commit verifier cost for a 50-validator quorum?",
			Model:          "EVM-native validator addresses, typed compact vote reconstruction, ecrecover per signer, >2/3 voting power.",
			Assumptions:    []string{"validators <= 256", "validator address = keccak256(uncompressed_pubkey[1:])[12:]", "signature digest = keccak256(compact vote bytes)", "recoverable signatures for ecrecover, not CometBFT's existing crypto/secp256k1 mode"},
			ExpectedOutput: "verifyCommitCompact returns true and consumes every bitmap-selected signature.",
		},
		{
			ID: "canonical-vote-reconstruct", Group: "canonical-vote", ExecutionRow: "reconstructCanonicalVoteSignBytes", CalldataRow: "canonical-vote-reconstruct", Fixture: "test/fixtures/update_50_equal.json",
			Question:       "What is the cost of reconstructing CometBFT CanonicalVote sign bytes from typed inputs?",
			Model:          "Solidity reconstructs the exact canonical vote protobuf bytes for a shared typed vote shape.",
			Assumptions:    []string{"vote type is PrecommitType", "block ID contains one PartSetHeader", "timestamp encoded as seconds+nanos protobuf Timestamp"},
			ExpectedOutput: "keccak256(reconstructed) equals keccak256(fixture canonical.vote.signBytes[i]) for every signer.",
		},
		{
			ID: "canonical-vote-reconstruct-hash", Group: "canonical-vote", ExecutionRow: "hashCanonicalVoteSignBytes", CalldataRow: "canonical-vote-reconstruct-hash", Fixture: "test/fixtures/update_50_equal.json",
			Question:       "What is the cost of reconstructing and hashing CometBFT CanonicalVote sign bytes?",
			Model:          "Solidity reconstructs canonical vote protobuf bytes and returns keccak256 of the reconstructed bytes.",
			Assumptions:    []string{"same typed vote shape as canonical-vote-reconstruct", "hash is keccak256 for the EVM secp256k1eth verifier path"},
			ExpectedOutput: "hashCanonicalVoteSignBytes equals keccak256(fixture canonical.vote.signBytes[i]).",
		},
		{
			ID: "canonical-vote-reconstruct-secp256k1", Group: "canonical-vote", ExecutionRow: "verifyCanonicalVoteSecp256k1", CalldataRow: "canonical-vote-reconstruct-secp256k1", Fixture: "test/fixtures/update_50_equal.json",
			Question:       "What is the cost of reconstructing canonical vote bytes, hashing them, and verifying one secp256k1eth signature?",
			Model:          "Solidity reconstructs canonical vote protobuf bytes, hashes them with keccak256, and verifies the fixture secp256k1eth signature with ecrecover.",
			Assumptions:    []string{"signature uses the EVM secp256k1eth fixture, not ed25519", "component row verifies one signer"},
			ExpectedOutput: "verifyCanonicalVoteSecp256k1 returns true for the fixture signer and signature.",
		},
		{
			ID: "ics23-wire-existence-decode-depth8", Group: "ICS23/IAVL", ExecutionRow: "decodeIcs23IavlExistenceProof", CalldataRow: "ics23-iavl-existence-decode-depth8", Fixture: "test/fixtures/update_50_equal.json",
			Question:       "What does decoding raw cosmos.ics23.v1.ExistenceProof calldata cost before verification?",
			Model:          "Wire-format IAVL existence proof known-field decoder for ExistenceProof, LeafOp, and repeated InnerOp.",
			Assumptions:    []string{"existence proof only", "unknown protobuf fields are skipped under proof-size and proof-depth caps", "happy-path-only gas is in ics23_happy_path_gas"},
			ExpectedOutput: "decoded proof depth and key hash equal the Go iavl/ics23 reference fixture.",
		},
		{
			ID: "ics23-wire-existence-verify-depth8", Group: "ICS23/IAVL", ExecutionRow: "verifyIcs23IavlExistenceProof(bytes32,bytes,bytes,bytes)", CalldataRow: "ics23-iavl-existence-verify-depth8", Fixture: "test/fixtures/update_50_equal.json",
			Question:       "What does raw cosmos.ics23.v1.ExistenceProof calldata decoding plus IAVL verification cost?",
			Model:          "Wire-format IAVL existence proof known-field decoder; validates SHA256 hash op, NO_HASH key prehash, SHA256 value prehash, VAR_PROTO length op, IAVL leaf prefix marker, and inner hash/suffix markers.",
			Assumptions:    []string{"existence proof only", "non-existence proof out of scope", "CommitmentProof wrapper reported separately only if added", "unknown protobuf fields are skipped under proof-size and proof-depth caps", "aggregate Forge rows include negative tests; happy-path-only gas is in ics23_happy_path_gas"},
			ExpectedOutput: "decoded proof root equals the Go iavl/ics23 reference root; wrong key/value/prefix/hash/length reject.",
		},
		{
			ID: "ics23-wire-existence-verify-depth16", Group: "ICS23/IAVL", ExecutionRow: "verifyIcs23IavlExistenceProof(bytes32,bytes,bytes,bytes)", CalldataRow: "ics23-iavl-existence-verify-depth16", Fixture: "test/fixtures/update_50_equal.json",
			Question:       "What does the maximum supported raw IAVL ExistenceProof depth fixture cost?",
			Model:          "Same wire-format IAVL existence proof known-field decoder and verifier as depth8, using the generated max-depth fixture.",
			Assumptions:    []string{"current generated maximum depth is 16", "contract cap is ICS23_MAX_DEPTH = 32", "non-existence proof out of scope"},
			ExpectedOutput: "decoded proof root equals the Go iavl/ics23 reference root for the max-depth fixture.",
		},
		{
			ID: "bls-a-supplied-aggregate", Group: "BLS-A", ExecutionRow: "verifyBlsAggregate", CalldataRow: "bls-a-aggregate-verify-supplied-aggregate-pubkey", Fixture: "test/fixtures/update_50_equal.json",
			Question:       "What does the smallest real EIP-2537 BLS aggregate verification cost when the aggregate pubkey, aggregate signature, and H(m) are supplied?",
			Model:          "Off-chain aggregate pubkey, off-chain aggregate signature, supplied H(m), one shared precommit message for the committed block.",
			Assumptions:    []string{"same-message BLS aggregation over one shared precommit message", "H(m) precomputed off-chain", "protocol changes CometBFT per-signer timestamped vote bytes into one shared precommit message", "production protocol needs BLS rogue-key mitigation such as proof-of-possession"},
			ExpectedOutput: "pairing equation e(-G1, sig) * e(aggPubKey, H(m)) == 1 returns true; wrong message/missing signer reject.",
		},
		{
			ID: "bls-b-stored-set", Group: "BLS-B", ExecutionRow: "verifyBlsAggregateStoredValidatorSet", CalldataRow: "bls-b-stored-validator-set-bitmap-50", Fixture: "test/fixtures/update_50_equal.json",
			Question:       "What does verifying a BLS aggregate cost when pubkeys and powers are stored on-chain by the canonical CometBFT BLS validator-set hash?",
			Model:          "Registration validates canonical 96-byte BLS pubkeys and powers against the fixture's canonical CometBFT BLS validator-set hash, then stores matching EIP-2537 G1 pubkeys and powers under that hash; verify calldata carries set hash, bitmap, threshold, aggregate signature, and H(m).",
			Assumptions:    []string{"one-time storage cost excluded from normal verify row", "storeBlsValidatorSet has its own execution and calldata row", "same-message BLS aggregation over one shared precommit message", "bitmap bits above the validator-set length reject", "production protocol needs BLS rogue-key mitigation such as proof-of-possession"},
			ExpectedOutput: "computed aggregate pubkey from stored signers verifies the aggregate signature and meets threshold power.",
		},
		{
			ID: "bls-b-store-validator-set", Group: "BLS-B", ExecutionRow: "storeBlsValidatorSet", CalldataRow: "bls-b-store-validator-set-50", Fixture: "test/fixtures/update_50_equal.json",
			Question:       "What is the one-time gas and calldata cost to store a 50-validator BLS validator set?",
			Model:          "Verify canonical 96-byte BLS pubkeys and powers against the supplied canonical CometBFT validator-set hash, then store matching EIP-2537 G1 pubkeys and powers under that hash.",
			Assumptions:    []string{"one-time setup/update cost, excluded from the normal verify row", "canonical and EIP-2537 pubkey arrays are the same validator order from the fixture", "production validator registration should verify or otherwise enforce BLS proof-of-possession"},
			ExpectedOutput: "storeBlsValidatorSet records the set under the supplied canonical hash, rejects a mismatched hash, and returns the canonical hash.",
		},
		{
			ID: "bls-c-calldata-set", Group: "BLS-C", ExecutionRow: "verifyBlsAggregateCalldataValidatorSet", CalldataRow: "bls-c-calldata-validator-set-50", Fixture: "test/fixtures/update_50_equal.json",
			Question:       "What is the cost when BLS validator pubkeys and powers are supplied in calldata and aggregated on-chain?",
			Model:          "Calldata validator pubkeys and powers, signer bitmap, on-chain aggregate pubkey, supplied aggregate signature and H(m).",
			Assumptions:    []string{"same-message BLS aggregation over one shared precommit message", "all pubkeys supplied as EIP-2537 G1 bytes", "calldata row matches the same ABI call as execution", "bitmap bits above the validator-set length reject", "production protocol needs BLS rogue-key mitigation such as proof-of-possession"},
			ExpectedOutput: "computed aggregate pubkey verifies the aggregate signature and threshold power.",
		},
		{
			ID: "bls-d-multi-message", Group: "BLS-D", ExecutionRow: "benchBlsMultiMessagePairing", Fixture: "synthetic EIP-2537 generator pairs",
			Question:       "Is per-signer multi-message BLS aggregate verification viable under EIP-2537 without collapsing votes to one shared precommit message?",
			Model:          "N signer-message pairings plus one aggregate-signature pairing, measured for N=10,50,100,175 and compared with the EIP-2537 formula.",
			Assumptions:    []string{"synthetic valid G1/G2 generator pairs measure precompile cost", "formula = 37,700 + 32,600 * pairCount with pairCount = N + 1", "does not prove signature validity; it prices the required N-pairing shape", "helper calldata rows encode only signerCount and are not production multi-message calldata"},
			ExpectedOutput: "benchmark returns false for N>1 but charges the N+1-pair precompile cost, proving non-viability by measured slope and formula.",
		},
	}
}

func parseGasRows(s string) []GasRow {
	var rows []GasRow
	for _, line := range strings.Split(s, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "|") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 7 {
			continue
		}
		name := strings.TrimSpace(parts[1])
		if name == "" || strings.Contains(name, "-") || name == "Function Name" || name == "Deployment Cost" {
			continue
		}
		min, ok := parseUintField(parts[2])
		if !ok {
			continue
		}
		avg, ok := parseUintField(parts[3])
		if !ok {
			continue
		}
		median, ok := parseUintField(parts[4])
		if !ok {
			continue
		}
		max, ok := parseUintField(parts[5])
		if !ok {
			continue
		}
		calls, ok := parseUintField(parts[6])
		if !ok {
			continue
		}
		rows = append(rows, GasRow{Function: name, Min: min, Avg: avg, Median: median, Max: max, Calls: calls})
	}
	return rows
}

func parseCalldataRows(s string) []CalldataRow {
	re := regexp.MustCompile(`CalldataCostBreakdown\(name: "([^"]+)".*byteLength: ([0-9]+).*zeroBytes: ([0-9]+).*nonzeroBytes: ([0-9]+).*gasL1Standard: ([0-9]+).*gasL1FloorEIP7623: ([0-9]+).*blobBytes: ([0-9]+)`)
	var rows []CalldataRow
	for _, line := range strings.Split(s, "\n") {
		m := re.FindStringSubmatch(line)
		if len(m) != 8 {
			continue
		}
		rows = append(rows, CalldataRow{
			Name:              m[1],
			ByteLength:        mustUint(m[2]),
			ZeroBytes:         mustUint(m[3]),
			NonzeroBytes:      mustUint(m[4]),
			GasL1Standard:     mustUint(m[5]),
			GasL1FloorEIP7623: mustUint(m[6]),
			BlobBytes:         mustUint(m[7]),
		})
	}
	return rows
}

func parseBlsD() []BlsDRow {
	var out []BlsDRow
	for _, n := range []uint64{10, 50, 100, 175} {
		rows := parseGasRows(readFile(fmt.Sprintf("generated/bls-d-%d.txt", n)))
		var measured uint64
		for _, row := range rows {
			if row.Function == "benchBlsMultiMessagePairing" {
				measured = row.Avg
				break
			}
		}
		pairings := n + 1
		formula := uint64(37700 + 32600*pairings)
		out = append(out, BlsDRow{
			Signers:       n,
			Pairings:      pairings,
			MeasuredGas:   measured,
			FormulaGas:    formula,
			OverheadGas:   int64(measured) - int64(formula),
			Formula:       "37_700 + 32_600 * (N + 1)",
			Source:        fmt.Sprintf("generated/bls-d-%d.txt", n),
			FoundryStable: measured > 0,
		})
	}
	return out
}

func parseIcs23HappyPath() []GasRow {
	files := []string{
		"generated/testIcs23IavlExistenceProofDecodeOnlyGas.txt",
		"generated/testIcs23IavlExistenceProofVerifyGas.txt",
	}
	var out []GasRow
	for _, file := range files {
		out = append(out, parseGasRows(readFile(file))...)
	}
	return out
}

func hashFixtures(dst map[string]string) {
	files, _ := filepath.Glob("test/fixtures/*.json")
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		sum := sha256.Sum256(data)
		dst[file] = hex.EncodeToString(sum[:])
	}
}

func writeTables(path string, m Manifest) {
	var b strings.Builder
	fmt.Fprintf(&b, "# Generated Benchmark Tables\n\n")
	fmt.Fprintf(&b, "Generated at `%s`.\n\n", m.GeneratedAt)
	fmt.Fprintf(&b, "## Toolchain\n\n")
	fmt.Fprintf(&b, "| Field | Value |\n|---|---|\n")
	fmt.Fprintf(&b, "| Forge | `%s` |\n", firstLine(m.Toolchain.Forge))
	fmt.Fprintf(&b, "| Solc | `%s` |\n", strings.Join(m.Toolchain.Solc, ", "))
	fmt.Fprintf(&b, "| Go | `%s` |\n", firstLine(m.Toolchain.Go))
	fmt.Fprintf(&b, "| Git commit | `%s` |\n", m.Toolchain.GitCommit)
	fmt.Fprintf(&b, "| Git status | `%s` |\n", oneLine(m.Toolchain.GitStatus))
	fmt.Fprintf(&b, "| EVM / optimizer / via IR | `%s / %t (%d runs) / %t` |\n\n", m.Toolchain.EVMVersion, m.Toolchain.Optimizer, m.Toolchain.OptimizerRuns, m.Toolchain.ViaIR)

	fmt.Fprintf(&b, "## BLS-D Multi-Message Pairing\n\n")
	fmt.Fprintf(&b, "| Signers N | Pairings | Measured gas | Formula gas | Overhead |\n|---:|---:|---:|---:|---:|\n")
	for _, row := range m.BlsDMultiMessage {
		fmt.Fprintf(&b, "| %d | %d | %d | %d | %+d |\n", row.Signers, row.Pairings, row.MeasuredGas, row.FormulaGas, row.OverheadGas)
	}
	fmt.Fprintf(&b, "\nFormula: `37_700 + 32_600 * (N + 1)` for EIP-2537 PAIRING_CHECK with `N` message/signature pairs plus the aggregate-signature pair.\n\n")
	fmt.Fprintf(&b, "The generated calldata rows named `bls-d-synthetic-helper-*` encode only the benchmark helper's `signerCount` and must not be used as production multi-message BLS calldata.\n\n")

	fmt.Fprintf(&b, "## ICS23 Happy-Path Execution Gas\n\n")
	fmt.Fprintf(&b, "| Function | Min | Avg | Median | Max | Calls |\n|---|---:|---:|---:|---:|---:|\n")
	for _, row := range m.Ics23HappyPath {
		fmt.Fprintf(&b, "| `%s` | %d | %d | %d | %d | %d |\n", row.Function, row.Min, row.Avg, row.Median, row.Max, row.Calls)
	}
	fmt.Fprintf(&b, "\n")

	fmt.Fprintf(&b, "## Selected Execution Gas\n\n")
	fmt.Fprintf(&b, "Aggregate Forge rows include all matching calls in the full test suite. For raw ICS23 happy-path-only decode/verify gas, use `benchmarks.json.ics23_happy_path_gas`.\n\n")
	fmt.Fprintf(&b, "| Function | Min | Avg | Median | Max | Calls |\n|---|---:|---:|---:|---:|---:|\n")
	interesting := map[string]bool{
		"verifyCommitCompact": true, "verifyCommitPrebuiltVoteBytes": true, "reconstructCanonicalVoteSignBytes": true,
		"hashCanonicalVoteSignBytes": true, "verifyCanonicalVoteSecp256k1": true, "decodeIcs23IavlExistenceProof": true,
		"verifyIcs23IavlExistenceProof(bytes32,bytes,bytes,bytes)": true, "verifyIavlExistenceProof": true,
		"verifyBlsAggregate": true, "storeBlsValidatorSet": true, "verifyBlsAggregateStoredValidatorSet": true,
		"verifyBlsAggregateCalldataValidatorSet": true, "benchBlsMultiMessagePairing": true,
		"benchBlsAggregateApprox": true, "verifyMembershipProof": true,
	}
	for _, row := range m.ExecutionGas {
		if interesting[row.Function] {
			fmt.Fprintf(&b, "| `%s` | %d | %d | %d | %d | %d |\n", row.Function, row.Min, row.Avg, row.Median, row.Max, row.Calls)
		}
	}

	fmt.Fprintf(&b, "\n## Calldata Rows\n\n")
	fmt.Fprintf(&b, "| Row | Bytes | Zero | Nonzero | Standard gas | EIP-7623 floor | Blob bytes |\n|---|---:|---:|---:|---:|---:|---:|\n")
	for _, row := range m.Calldata {
		fmt.Fprintf(&b, "| `%s` | %d | %d | %d | %d | %d | %d |\n", row.Name, row.ByteLength, row.ZeroBytes, row.NonzeroBytes, row.GasL1Standard, row.GasL1FloorEIP7623, row.BlobBytes)
	}
	_ = os.WriteFile(path, []byte(b.String()), 0o644)
}

func writeJSON(path string, v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		panic(err)
	}
}

func commandOutput(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return strings.TrimSpace(out.String())
	}
	return strings.TrimSpace(out.String())
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func parseUintField(s string) (uint64, bool) {
	fields := strings.Fields(strings.TrimSpace(s))
	if len(fields) == 0 {
		return 0, false
	}
	v, err := strconv.ParseUint(fields[0], 10, 64)
	return v, err == nil
}

func mustUint(s string) uint64 {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func parseSolcVersions(s string) []string {
	re := regexp.MustCompile(`Solc ([0-9]+\.[0-9]+\.[0-9]+)`)
	seen := map[string]bool{}
	var versions []string
	for _, match := range re.FindAllStringSubmatch(s, -1) {
		if !seen[match[1]] {
			seen[match[1]] = true
			versions = append(versions, match[1])
		}
	}
	if len(versions) == 0 {
		return []string{"source pragmas =0.8.30, =0.6.12"}
	}
	return versions
}

func jsonString(s, key string) string {
	var v map[string]any
	if json.Unmarshal([]byte(s), &v) != nil {
		return ""
	}
	if x, ok := v[key].(string); ok {
		return x
	}
	return ""
}

func jsonBool(s, key string) bool {
	var v map[string]any
	if json.Unmarshal([]byte(s), &v) != nil {
		return false
	}
	x, _ := v[key].(bool)
	return x
}

func jsonUint(s, key string) uint64 {
	var v map[string]any
	if json.Unmarshal([]byte(s), &v) != nil {
		return 0
	}
	switch x := v[key].(type) {
	case float64:
		return uint64(x)
	case json.Number:
		u, _ := strconv.ParseUint(string(x), 10, 64)
		return u
	default:
		return 0
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func oneLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "clean"
	}
	return strings.ReplaceAll(s, "\n", "; ")
}
