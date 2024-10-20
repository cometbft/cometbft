package core

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
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
	// The test genesis file is the genesis that the ci.toml e2e test uses.
	t.Run("NoChunking", func(t *testing.T) {
		const fGenesisPath = "./testdata/genesis_ci.json"

		genesisData, err := os.ReadFile(fGenesisPath)
		if err != nil {
			t.Fatalf("reading test genesis file at %q: %s", fGenesisPath, err)
		}

		genDoc := &types.GenesisDoc{}
		if err := cmtjson.Unmarshal(genesisData, genDoc); err != nil {
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
	// The test genesis file has an app_state of key-value string pairs
	// automatically generated (~42MB).
	t.Run("Chunking", func(t *testing.T) {
		const fGenesisPath = "./testdata/genesis_big.json"

		genesisData, err := os.ReadFile(fGenesisPath)
		if err != nil {
			t.Fatalf("reading test genesis file at %q: %s", fGenesisPath, err)
		}

		genesisDoc := &types.GenesisDoc{}
		if err := cmtjson.Unmarshal(genesisData, genesisDoc); err != nil {
			t.Fatalf("test genesis serialization: %s", err)
		}

		env := &Environment{
			genChunks: nil,
			GenDoc:    genesisDoc,
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
			genesisJSON, err := cmtjson.Marshal(genesisDoc)
			if err != nil {
				t.Fatalf("test genesis re-serialization: %s", err)
			}

			// Because the genesis file is > genesisChunkSize, we expect chunks.
			// genesisChunkSize is a global const defined in env.go.
			genesisSize := len(genesisJSON)
			wantChunks := (genesisSize + genesisChunkSize - 1) / genesisChunkSize
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

func TestDeleteGenesisChunks(t *testing.T) {
	t.Run("NoErrDirNotExist", func(t *testing.T) {
		env := &Environment{GenesisFilePath: "/nonexistent/path/to/genesis.json"}

		if err := env.deleteGenesisChunks(); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})

	t.Run("DirDeleted", func(t *testing.T) {
		gFileDir, err := os.MkdirTemp("", "test_dir")
		if err != nil {
			t.Fatalf("creating temp directory for testing: %s", err)
		}
		defer os.RemoveAll(gFileDir)

		var (
			gFilePath = filepath.Join(gFileDir, "genesis.json")
			chunksDir = filepath.Join(gFileDir, _chunksDir)

			env = &Environment{GenesisFilePath: gFilePath}
		)

		// the directory we want to delete.
		if err := os.MkdirAll(chunksDir, 0o755); err != nil {
			t.Fatalf("creating test chunks directory: %s", err)
		}

		if err := env.deleteGenesisChunks(); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		// verify that chunksDir no longer exists
		if _, err := os.Stat(chunksDir); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("expected os.IsNotExist error, but got: %s", err)
		}
	})

	t.Run("ErrAccessingDirPath", func(t *testing.T) {
		var (
			// To test if the function catches errors returns by os.Stat() that
			// aren't fs.ErrNotExist, we create a path that contains an invalid null
			// byte, thus forcing os.Stat() to return an error.
			gFilePath = "null/" + string('\x00') + "/path"
			env       = &Environment{GenesisFilePath: gFilePath}
		)

		err := env.deleteGenesisChunks()
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}

		wantErr := "accessing path \"null/\\x00/chunks\": stat null/\x00/chunks: invalid argument"
		if err.Error() != wantErr {
			t.Errorf("\nwant error: %s\ngot: %s\n", wantErr, err.Error())
		}
	})

	t.Run("ErrDeletingDir", func(t *testing.T) {
		// To test if the function catches errors returned by os.RemoveAll(), we
		// create a directory with read-only permissions, so that os.RemoveAll() will
		// fail.
		// Usually, the deletion of a file or a directory is controlled by the
		// permissions of the *parent* directory. Therefore, in this test we are
		// creating a directory and a sub-directory; then we'll set the parent
		// directory's permissions to read-only, so that os.RemoveAll() will fail.

		parentDir, err := os.MkdirTemp("", "parentDir")
		if err != nil {
			t.Fatalf("creating test parent directory: %s", err)
		}
		defer os.RemoveAll(parentDir)

		var (
			gFilePath = filepath.Join(parentDir, "genesis.json")
			chunksDir = filepath.Join(parentDir, _chunksDir)

			env = &Environment{GenesisFilePath: gFilePath}
		)

		// the sub-directory that we want to delete.
		if err := os.Mkdir(chunksDir, 0o755); err != nil {
			t.Fatalf("creating test chunks directory: %s", err)
		}

		// set read-only permissions to trigger deletion error
		if err := os.Chmod(parentDir, 0o500); err != nil {
			t.Fatalf("changing test parent directory permissions: %s", err)
		}

		err = env.deleteGenesisChunks()
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}

		var (
			formatStr = "deleting pre-existing genesis chunks at %s: unlinkat %s: permission denied"
			wantErr   = fmt.Sprintf(formatStr, chunksDir, chunksDir)
		)
		if err.Error() != wantErr {
			t.Errorf("\nwant error: %s\ngot: %s\n", wantErr, err.Error())
		}

		// reset permissions of parent folder to allow the deferred os.RemoveAll()
		// to work, thus deleting test data.
		if err := os.Chmod(parentDir, 0o755); err != nil {
			formatStr := "changing test parent directory permissions to cleanup: %s"
			t.Fatalf(formatStr, err)
		}
	})
}

func TestFileSize(t *testing.T) {
	t.Run("ErrFileNotExist", func(t *testing.T) {
		fPath := "non-existent-file"
		_, err := fileSize(fPath)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}

		wantErr := "the file is unavailable at non-existent-file"
		if err.Error() != wantErr {
			t.Fatalf("\nwant error: %s\ngot: %s\n", wantErr, err.Error())
		}
	})

	t.Run("ErrAccessingPath", func(t *testing.T) {
		// To test if the function catches errors returns by os.Stat() that
		// aren't fs.ErrNotExist, we create a path that contains an invalid null
		// byte, thus forcing os.Stat() to return an error.
		fPath := "null/" + string('\x00') + "/file"

		_, err := fileSize(fPath)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}

		wantErr := "accessing file at null/\x00/file: stat null/\x00/file: invalid argument"
		if err.Error() != wantErr {
			t.Errorf("\nwant error: %s\ngot: %s\n", wantErr, err.Error())
		}
	})

	t.Run("FileSizeOk", func(t *testing.T) {
		// we'll create a temporary file of 100 bytes to run this test.
		const fTempSize = 100

		fTemp, err := os.CreateTemp("", "small_test_file")
		if err != nil {
			t.Fatalf("creating temp file for testing: %s", err)
		}
		defer os.Remove(fTemp.Name())

		data := make([]byte, fTempSize)
		for i := 0; i < 100; i++ {
			data[i] = 'a'
		}

		if _, err := fTemp.Write(data); err != nil {
			t.Fatalf("writing to temp file for testing: %s", err)
		}
		fTemp.Close()

		gotSize, err := fileSize(fTemp.Name())
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if gotSize != fTempSize {
			t.Errorf("want size: %d, got: %d", fTempSize, gotSize)
		}
	})
}

func TestMkChunksDir(t *testing.T) {
	// TODO: test that makes MkDir fail to check error handling.

	t.Run("DirCreated", func(t *testing.T) {
		fTemp, err := os.CreateTemp("", "dummy_genesis")
		if err != nil {
			t.Fatalf("creating temp file for testing: %s", err)
		}
		fTemp.Close()
		defer os.Remove(fTemp.Name())

		dirPath, err := mkChunksDir(fTemp.Name(), _chunksDir)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		if err := os.RemoveAll(dirPath); err != nil {
			t.Error(err)
		}
	})
}

func TestWriteChunk(t *testing.T) {
	cDir, err := os.MkdirTemp("", _chunksDir)
	if err != nil {
		t.Fatalf("creating test chunks directory: %s", err)
	}
	defer os.RemoveAll(cDir)

	var (
		chunk    = []byte("test-chunk")
		wantPath = filepath.Join(cDir, "chunk_42.part")
	)

	t.Run("ErrChunkNotWritten", func(t *testing.T) {
		// To test if the function catches errors returned by os.WriteFile(), we
		// create a directory with read-only permissions, so that os.WriteFile() will
		// fail.

		// set read-only permissions to trigger write error
		if err := os.Chmod(cDir, 0o500); err != nil {
			t.Fatalf("changing test chunks directory permissions: %s", err)
		}

		_, err := writeChunk(chunk, cDir, 42)
		if err == nil {
			t.Fatalf("expected error but got nil")
		}

		var (
			formatStr = "writing chunk at %s: open %s: permission denied"
			wantErr   = fmt.Sprintf(formatStr, wantPath, wantPath)
		)
		// reset permissions of chunks folder to allow the rest of the test code to
		// work.
		if err := os.Chmod(cDir, 0o755); err != nil {
			formatStr := "changing test parent directory permissions to cleanup: %s"
			t.Fatalf(formatStr, err)
		}
		if err.Error() != wantErr {
			t.Errorf("\nwant error: %s\ngot: %s\n", wantErr, err.Error())
		}
	})

	t.Run("ChunkWritten", func(t *testing.T) {
		gotPath, err := writeChunk(chunk, cDir, 42)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if wantPath != gotPath {
			t.Errorf("\nwant path: %s\ngot path: %s\n", wantPath, gotPath)
		}
	})

}
