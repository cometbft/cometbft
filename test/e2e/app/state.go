package app

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

const (
	stateFileName     = "app_state.json"
	prevStateFileName = "prev_app_state.json"
)

// Used exclusively in serialization/deserialization of State.
type serializedState struct {
	Height uint64
	Values map[string]string
	Hash   []byte
}

// State is the application state.
type State struct {
	sync.RWMutex
	height uint64
	values map[string]string
	hash   []byte

	// private fields aren't marshaled to disk.
	currentFile string
	// app saves current and previous state for rollback functionality
	previousFile    string
	persistInterval uint64
	initialHeight   uint64
}

// NewState creates a new state.
func NewState(dir string, persistInterval uint64) (*State, error) {
	state := &State{
		values:          make(map[string]string),
		currentFile:     filepath.Join(dir, stateFileName),
		previousFile:    filepath.Join(dir, prevStateFileName),
		persistInterval: persistInterval,
	}
	state.hash = hashItems(state.values, state.height)
	err := state.load()
	switch {
	case errors.Is(err, os.ErrNotExist):
	case err != nil:
		return nil, err
	}
	return state, nil
}

// load loads state from disk. It does not take out a lock, since it is called
// during construction.
func (s *State) load() error {
	bz, err := os.ReadFile(s.currentFile)
	if err != nil {
		// if the current state doesn't exist then we try recover from the previous state
		if errors.Is(err, os.ErrNotExist) {
			bz, err = os.ReadFile(s.previousFile)
			if err != nil {
				return fmt.Errorf("failed to read both current and previous state (%q): %w",
					s.previousFile, err)
			}
		} else {
			return fmt.Errorf("failed to read state from %q: %w", s.currentFile, err)
		}
	}
	var ss serializedState
	err = json.Unmarshal(bz, &ss)
	if err != nil {
		return fmt.Errorf("invalid state data in %q: %w", s.currentFile, err)
	}
	s.height = ss.Height
	s.values = ss.Values
	s.hash = ss.Hash
	return nil
}

// save saves the state to disk. It does not take out a lock since it is called
// internally by Commit which does lock.
func (s *State) save() error {
	ss := serializedState{
		Height: s.height,
		Values: s.values,
		Hash:   s.hash,
	}
	bz, err := json.Marshal(&ss)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}
	// We write the state to a separate file and move it to the destination, to
	// make it atomic.
	newFile := fmt.Sprintf("%v.new", s.currentFile)
	err = os.WriteFile(newFile, bz, 0o644) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to write state to %q: %w", s.currentFile, err)
	}
	// We take the current state and move it to the previous state, replacing it
	if _, err := os.Stat(s.currentFile); err == nil {
		if err := os.Rename(s.currentFile, s.previousFile); err != nil {
			return fmt.Errorf("failed to replace previous state: %w", err)
		}
	}
	// Finally, we take the new state and replace the current state.
	return os.Rename(newFile, s.currentFile)
}

// GetHeight provides a thread-safe way of accessing the current height of the
// state.
func (s *State) GetHeight() uint64 {
	s.RLock()
	defer s.RUnlock()
	return s.height
}

// GetHash provides a thread-safe way of accessing a copy of the current state
// hash.
func (s *State) GetHash() []byte {
	s.RLock()
	defer s.RUnlock()
	hash := make([]byte, len(s.hash))
	copy(hash, s.hash)
	return hash
}

// GetValues provides a thread-safe way of obtaining a copy of the current
// state values.
func (s *State) GetValues() map[string]string {
	s.RLock()
	defer s.RUnlock()
	values := make(map[string]string, len(s.values))
	for k, v := range s.values {
		values[k] = v
	}
	return values
}

// Export exports key/value pairs as JSON, used for state sync snapshots.
func (s *State) Export() ([]byte, error) {
	s.RLock()
	defer s.RUnlock()
	return json.Marshal(s.values)
}

// Import imports key/value pairs from JSON bytes, used for InitChain.AppStateBytes and
// state sync snapshots. It also saves the state once imported.
func (s *State) Import(height uint64, jsonBytes []byte) error {
	s.Lock()
	defer s.Unlock()
	values := map[string]string{}
	err := json.Unmarshal(jsonBytes, &values)
	if err != nil {
		return fmt.Errorf("failed to decode imported JSON data: %w", err)
	}
	s.height = height
	s.values = values
	s.hash = hashItems(values, height)
	return s.save()
}

// Get fetches a value. A missing value is returned as an empty string.
func (s *State) Get(key string) string {
	s.RLock()
	defer s.RUnlock()
	return s.values[key]
}

// Set sets a value. Setting an empty value is equivalent to deleting it.
func (s *State) Set(key, value string) {
	s.Lock()
	defer s.Unlock()
	if value == "" {
		delete(s.values, key)
	} else {
		s.values[key] = value
	}
}

// Finalize is called after applying a block, updating the height and returning the new app_hash
func (s *State) Finalize() []byte {
	s.Lock()
	defer s.Unlock()
	switch {
	case s.height > 0:
		s.height++
	case s.initialHeight > 0:
		s.height = s.initialHeight
	default:
		s.height = 1
	}
	s.hash = hashItems(s.values, s.height)
	return s.hash
}

// Commit commits the current state.
func (s *State) Commit() (uint64, error) {
	s.Lock()
	defer s.Unlock()
	if s.persistInterval > 0 && s.height%s.persistInterval == 0 {
		err := s.save()
		if err != nil {
			return 0, err
		}
	}
	return s.height, nil
}

func (s *State) Rollback() error {
	bz, err := os.ReadFile(s.previousFile)
	if err != nil {
		return fmt.Errorf("failed to read state from %q: %w", s.previousFile, err)
	}
	var ss serializedState
	err = json.Unmarshal(bz, &ss)
	if err != nil {
		return fmt.Errorf("invalid state data in %q: %w", s.previousFile, err)
	}
	s.height = ss.Height
	s.hash = ss.Hash
	s.values = ss.Values
	return nil
}

// hashItems hashes a set of key/value items.
func hashItems(items map[string]string, height uint64) []byte {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	hasher := sha256.New()
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], height)
	_, _ = hasher.Write(b[:])
	for _, key := range keys {
		_, _ = hasher.Write([]byte(key))
		_, _ = hasher.Write([]byte{0})
		_, _ = hasher.Write([]byte(items[key]))
		_, _ = hasher.Write([]byte{0})
	}
	return hasher.Sum(nil)
}
