package v2

import (
    "fmt"
)
// SafeRandUnrecognizedParams wraps randUnrecognizedParams to prevent integer overflow
// by capping maxFieldNumber at 2^29 - 1 (536870911). This ensures fieldNumber stays
// within the valid Protocol Buffers range [1, 536870911], avoiding uint32 overflow
// in the key calculation within randFieldParams.
func SafeRandUnrecognizedParams(r randyParams, maxFieldNumber int) ([]byte, error) {
    if maxFieldNumber > 536870911 {
        return nil, fmt.Errorf("maxFieldNumber %d exceeds maximum (536870911)", maxFieldNumber)
    }
    return randUnrecognizedParams(r, maxFieldNumber), nil
}