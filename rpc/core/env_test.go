package core

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/types"
)

func TestInitGenesisChunks(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		env := &Environment{
			genChunks: nil,
			GenDoc:    nil,
		}
		wantErrStr := "could not create the genesis file chunks and cache them because the genesis doc is unavailable"

		if err := env.InitGenesisChunks(); err == nil {
			t.Error("expected error but got nil")
		} else if err.Error() != wantErrStr {
			t.Errorf("\nwantErr: %q\ngot: %q\n", wantErrStr, err.Error())
		}
	})

	// Calling InitGenesisChunks with an existing slice of chunks will return without
	// doing anything.
	t.Run("NoOp", func(t *testing.T) {
		testChunks := []string{"chunk1", "chunk2"}
		env := &Environment{
			genChunks: testChunks,
			GenDoc:    nil,
		}

		if err := env.InitGenesisChunks(); err != nil {
			t.Errorf("unexpected error: %s", err)
		} else {
			if !slices.Equal(testChunks, env.genChunks) {
				t.Fatalf("\nexpected chunks: %v\ngot: %v", testChunks, env.genChunks)
			}
			if env.GenDoc != nil {
				formatStr := "pointer to GenesisDoc should be nil, but it's pointing to\n%#v"
				t.Errorf(formatStr, *env.GenDoc)
			}
		}
	})

	// Tests with a genesis file <= genesisChunkSize, i.e., no chunking, pointer to
	// GenesisDoc stored in GenDoc field.
	// The test genesis is the genesis that the ci.toml e2e test uses.
	t.Run("NoChunking", func(t *testing.T) {
		genDoc := &types.GenesisDoc{}
		if err := cmtjson.Unmarshal([]byte(_testGenesis), genDoc); err != nil {
			t.Fatalf("test genesis serialization: %s", err)
		}

		env := &Environment{
			genChunks: nil,
			GenDoc:    genDoc,
		}

		if err := env.InitGenesisChunks(); err != nil {
			t.Errorf("unexpected error: %s", err)
		} else {
			if env.genChunks != nil {
				formatStr := "chunks slice should be nil, but it has length %d"
				t.Fatalf(formatStr, len(env.genChunks))
			}

			// Because the genesis file is <= genesisChunkSize, there should be no
			// chunking. Therefore, the original GenesisDoc should be stored in
			// GenDoc field unchanged.
			if !reflect.DeepEqual(env.GenDoc, genDoc) {
				formatStr := "GenesisDoc in Environment.GenDoc should be the same as in test genesis file\nwant: %#v\ngot: %#v\n"
				t.Errorf(formatStr, genDoc, env.GenDoc)
			}
		}
	})

	// Tests with a genesis file > genesisChunkSize, i.e., chunking, pointer to
	// GenesisDoc is nil, chunks slice stored in genChunks field.
	// The test genesis has an app_state of key-value string pairs automatically
	// generated (~42MB).
	t.Run("Chunking", func(t *testing.T) {
		genDoc := &types.GenesisDoc{}
		if err := cmtjson.Unmarshal([]byte(_testGenesis), genDoc); err != nil {
			t.Fatalf("test genesis serialization: %s", err)
		}

		appState, err := genAppState()
		if err != nil {
			t.Fatalf("generating dummy app_state for testing: %s", err)
		}

		genDoc.AppState = appState

		env := &Environment{
			genChunks: nil,
			GenDoc:    genDoc,
		}

		if err := env.InitGenesisChunks(); err != nil {
			t.Errorf("unexpected error: %s", err)
		} else {
			if env.GenDoc != nil {
				formatStr := "pointer to GenesisDoc should be nil, but it's pointing to\n%#v"
				t.Fatalf(formatStr, *env.GenDoc)
			}

			// Why do we re-marshal the genesis to JSON?
			// Because InitGenesisChunks computes the number of chunks based on the
			// size of the []byte slice containing the genesis serialized to JSON.
			// To calculate the correct expected number of chunks in this test, we
			// must also serialize the genesis to JSON and use the size of the
			// resulting []byte slice.
			// We cannot use the size of the []byte slice obtained from reading the
			// file (`genesisData` above) because the size would differ due to JSON
			// serialization removing whitespace and formatting, omitting default or
			// zero values, and optimizing data (e.g., numbers).
			genesisJSON, err := cmtjson.Marshal(genDoc)
			if err != nil {
				t.Fatalf("test genesis re-serialization: %s", err)
			}

			// Because the genesis file is > genesisChunkSize, we expect chunks.
			// genesisChunkSize is a global const defined in env.go.
			var (
				genesisSize = len(genesisJSON)
				wantChunks  = (genesisSize + genesisChunkSize - 1) / genesisChunkSize
			)
			if len(env.genChunks) != wantChunks {
				formatStr := "expected number of chunks: %d, but got: %d"
				t.Errorf(formatStr, wantChunks, len(env.genChunks))
			}

			// We now check if the original genesis doc and the genesis doc
			// reassembled from the chunks match.
			var genesisReassembled bytes.Buffer
			for i, chunk := range env.genChunks {
				chunkBytes, err := base64.StdEncoding.DecodeString(chunk)
				if err != nil {
					t.Fatalf("failed to decode chunk %d: %s", i, err)
				}

				if _, err := genesisReassembled.Write(chunkBytes); err != nil {
					t.Fatalf("failed to write chunk %d to buffer: %s", i, err)
				}
			}

			if !bytes.Equal(genesisReassembled.Bytes(), genesisJSON) {
				t.Errorf("original and reassembled genesis do not match")
			}
		}
	})
}

func TestPaginationPage(t *testing.T) {
	cases := []struct {
		totalCount int
		perPage    int
		page       int
		newPage    int
		expErr     bool
	}{
		{0, 10, 1, 1, false},

		{0, 10, 0, 1, false},
		{0, 10, 1, 1, false},
		{0, 10, 2, 0, true},

		{5, 10, -1, 0, true},
		{5, 10, 0, 1, false},
		{5, 10, 1, 1, false},
		{5, 10, 2, 0, true},
		{5, 10, 2, 0, true},

		{5, 5, 1, 1, false},
		{5, 5, 2, 0, true},
		{5, 5, 3, 0, true},

		{5, 3, 2, 2, false},
		{5, 3, 3, 0, true},

		{5, 2, 2, 2, false},
		{5, 2, 3, 3, false},
		{5, 2, 4, 0, true},
	}

	for _, c := range cases {
		p, err := validatePage(&c.page, c.perPage, c.totalCount)
		if c.expErr {
			require.Error(t, err)
			continue
		}

		assert.Equal(t, c.newPage, p, fmt.Sprintf("%v", c))
	}

	// nil case
	p, err := validatePage(nil, 1, 1)
	if assert.NoError(t, err) { //nolint:testifylint // require.Error doesn't work with the conditional here
		assert.Equal(t, 1, p)
	}
}

func TestPaginationPerPage(t *testing.T) {
	cases := []struct {
		totalCount int
		perPage    int
		newPerPage int
	}{
		{5, 0, defaultPerPage},
		{5, 1, 1},
		{5, 2, 2},
		{5, defaultPerPage, defaultPerPage},
		{5, maxPerPage - 1, maxPerPage - 1},
		{5, maxPerPage, maxPerPage},
		{5, maxPerPage + 1, maxPerPage},
	}
	env := &Environment{}
	for _, c := range cases {
		p := env.validatePerPage(&c.perPage)
		assert.Equal(t, c.newPerPage, p, fmt.Sprintf("%v", c))
	}

	// nil case
	p := env.validatePerPage(nil)
	assert.Equal(t, defaultPerPage, p)
}

// genAppState is a helper function that generates a dummy "app_state" to be used in
// tests. To test the splitting of a genesis into smaller chunks, we need to use a
// big genesis file. Typically, the bulk of a genesis file comes from the app_state
// field.
// It returns the app_state encoded to JSON.
func genAppState() ([]byte, error) {
	const (
		// how many KV pair do you want to put in app_state.
		// Current value generates an app_state of ~40MB
		size = 1024 * 1024 * 2

		// characters use to fill in the KV pairs of app_state
		alphabet     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		alphabetSize = len(alphabet)
	)

	appState := make(map[string]string, size)
	for i := range size {
		appState["initial"+strconv.Itoa(i)] = string(alphabet[i%alphabetSize])
	}

	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		return nil, fmt.Errorf("serializing test app_state to JSON: %s", err)
	}

	return appStateJSON, nil
}
