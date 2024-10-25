package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/types"
)

func TestInitGenesisChunks(t *testing.T) {
	t.Run("ErrNoGenesisFilePath", func(t *testing.T) {
		env := &Environment{}

		err := env.InitGenesisChunks()
		if err == nil {
			t.Error("expected error but got nil")
		}

		wantErrStr := "missing genesis file path on disk"
		if err.Error() != wantErrStr {
			t.Errorf("\nwantErr: %q\ngot: %q\n", wantErrStr, err.Error())
		}
	})

	// Calling InitGenesisChunks with an existing slice of chunks will return without
	// doing anything.
	t.Run("NoOp", func(t *testing.T) {
		testChunks := map[int]string{
			0: "chunk1",
			1: "chunk2",
		}
		env := &Environment{genesisChunks: testChunks}

		err := env.InitGenesisChunks()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		// check that the function really was a no-op: the map of chunks should be
		// unchanged.
		if !maps.Equal(testChunks, env.genesisChunks) {
			t.Fatalf("\nexpected chunks: %v\ngot: %v", testChunks, env.genesisChunks)
		}
	})

	// Tests with a genesis file <= genesisChunkSize, i.e., no chunking.
	// The test genesis file is the genesis that the ci.toml e2e test uses.
	t.Run("NoChunking", func(t *testing.T) {
    fGenesis, err := os.CreateTemp("", "genesis.json")
	  if err != nil {
      t.Fatalf("creating genesis file for testing: %s", err)
	  }
	  defer os.Remove(fTemp.Name())
    
    if _, err := fGenesis.Write([]byte(_testGenesis)); err != nil {
      t.Fatalf("writing genesis file for testing: %s", err)
	  }
	  fGenesis.Close()
		
    env := &Environment{GenesisFilePath: fGenesis.Name()}

		err := env.InitGenesisChunks()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		// Because the genesis file is <= genesisChunkSize, there should be no
		// chunking. Therefore, the map of chunk IDs to their paths on disk should
		// be empty.
		if len(env.genesisChunks) > 0 {
			formatStr := "chunks map should be empty, but it's %v"
			t.Fatalf(formatStr, env.genesisChunks)
    }
  })

	// Tests with a genesis file > genesisChunkSize.
	// The test genesis file has an app_state of key-value string pairs
	// automatically generated (~42MB).
	t.Run("Chunking", func(t *testing.T) {
    genDoc := &types.GenesisDoc{}
		if err := cmtjson.Unmarshal([]byte(_testGenesis), genDoc); err != nil {
			t.Fatalf("test genesis de-serialization: %s", err)
		}

		appState, err := genAppState()
		if err != nil {
			t.Fatalf("generating dummy app_state for testing: %s", err)
		}

		genDoc.AppState = appState
    
    genDocJSON, err := cmtjson.Marshal(genDoc)
    if err != nil {
      t.Fatalf("test genesis serialization: %s", err)
    }
    
    fGenesis, err := os.CreateTemp("", "genesis.json")
	  if err != nil {
      t.Fatalf("creating genesis file for testing: %s", err)
	  }
    
    if _, err := fGenesis.Write(genDocJSON); err != nil {
      t.Fatalf("writing genesis file for testing: %s", err)
	  }
	  fGenesis.Close()
    
    var (
      fGenesisPath := filepath.Join(filepath.Dir(fGenesis.Name()), _chunksDir)
      env := &Environment{GenesisFilePath: fGenesisPath}
    )
		defer os.RemoveAll(fGenesisPath)

		err := env.InitGenesisChunks()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		genesisSize, err := fileSize(fGenesisPath)
		if err != nil {
			t.Fatalf("estimating test genesis file size: %s", err)
		}

		// Because the genesis file is > genesisChunkSize, we expect chunks.
		// genesisChunkSize is a global const defined in env.go.
		wantChunks := (genesisSize + genesisChunkSize - 1) / genesisChunkSize
		if len(env.genesisChunks) != wantChunks {
			formatStr := "expected number of chunks: %d, but got: %d"
			t.Errorf(formatStr, wantChunks, len(env.genesisChunks))
		}

		// We now check if the original genesis doc and the genesis doc
		// reassembled from the chunks match.
		if err := reassembleAndCompare(fGenesisPath, env.genesisChunks); err != nil {
			t.Errorf("reassembling genesis file: %s", err)
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
		var (
			gFilePath = "./genesis.json"
			chunksDir = "./" + _chunksDir

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
			formatStr = "deleting genesis chunks' folder at %s: unlinkat %s: permission denied"
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

func TestWriteChunks(t *testing.T) {
	const (
		// we'll create a temporary file of 100 bytes to run this test.
		fTempSize = 100

		// we'll split the temp file into chunks of 25 bytes.
		chunkSize = 25
	)

	fTemp, err := os.CreateTemp("", "dummy_genesis")
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

	chunkIDToPath, err := writeChunks(fTemp.Name(), chunkSize)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	var (
		fTempDir      = filepath.Dir(fTemp.Name())
		testChunksDir = filepath.Join(fTempDir, _chunksDir)
	)
	defer os.RemoveAll(testChunksDir)

	wantMap := map[int]string{
		0: testChunksDir + "/chunk_0.part",
		1: testChunksDir + "/chunk_1.part",
		2: testChunksDir + "/chunk_2.part",
		3: testChunksDir + "/chunk_3.part",
	}
	if !maps.Equal(wantMap, chunkIDToPath) {
		t.Errorf("\nwant map: %v\ngot: %v\n", wantMap, chunkIDToPath)
	}

	if err := reassembleAndCompare(fTemp.Name(), chunkIDToPath); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

// reassembleAndCompare is a helper function to reassemble the genesis file from
// its chunks and compare it with the original genesis file.
// The function reads the genesis file as a stream, so it is suitable for larger
// files as well.
func reassembleAndCompare(gFilePath string, chunks map[int]string) error {
	gFile, err := os.Open(gFilePath)
	if err != nil {
		return fmt.Errorf("opening genesis file at %s: %s", gFilePath, err)
	}
	defer gFile.Close()

	// have to collect the IDs and sort them because map traversal isn't guaranteed
	// to be in order; but we need it to be in order to compare the right chunk
	// with the right of the genesis file "piece".
	cIDs := make([]int, 0, len(chunks))
	for cID := range chunks {
		cIDs = append(cIDs, cID)
	}
	slices.Sort(cIDs)

	for cID := range cIDs {
		cPath := chunks[cID]

		chunk, err := os.ReadFile(cPath)
		if err != nil {
			return fmt.Errorf("reading chunk file %d: %s", cID, err)
		}

		gBuf := make([]byte, len(chunk))
		gN, err := gFile.Read(gBuf)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("reading genesis file chunk %d: %s", cID, err)
		}

		if !bytes.Equal(gBuf[:gN], chunk) {
			return fmt.Errorf("chunk %d does not match", cID)
		}
	}

	return nil
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
