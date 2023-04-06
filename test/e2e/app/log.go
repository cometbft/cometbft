package app

import (
	"fmt"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
)

const ABCI_CALL = "abci"

// GetRequestString gets the string representation of the request that will be logged by the application.
func GetRequestString(req *abci.Request) (string, error) {
	b, err := proto.Marshal(req)
	if err != nil {
		return "", err
	}
	s := ABCI_CALL + "|" + string(b)
	return s, nil
}

// GetRequestFromString gets a Request from a string created by GetRequestString.
func GetRequestFromString(s string) (*abci.Request, error) {
	parts := strings.Split(s, "|")
	if len(parts) != 2 || parts[0] != ABCI_CALL {
		return nil, fmt.Errorf("String passed to GetRequestFromString does not have a good format!\n")
	}
	req := &abci.Request{}
	reqBytes := []byte(parts[1])
	err := proto.Unmarshal(reqBytes, req)
	if err != nil {
		return nil, err
	}
	return req, nil
}
