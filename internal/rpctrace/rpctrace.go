package rpctrace

import "github.com/gofrs/uuid"

// New returns a randomly generated string which can be used to assist in
// tracing RPC errors.
func New() (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
