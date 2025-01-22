package trace

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
)

// DecodeFile reads a file and decodes it into a slice of events via
// scanning. The table parameter is used to determine the type of the events.
// The file should be a jsonl file. The generic here are passed to the event
// type.
func DecodeFile[T any](f *os.File) ([]Event[T], error) {
	var out []Event[T]
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		var e Event[T]
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, err
		}

		out = append(out, e)
	}

	return out, nil
}
