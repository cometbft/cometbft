package app

import (
	"fmt"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
)

const ABCI_REQ = "abci"

// GetRequestString gets the string representation of the request that will be logged by the application.
func GetRequestString(req *abci.Request) (string, error) {
	b, err := proto.Marshal(req)
	if err != nil {
		return "", err
	}
	s := ABCI_REQ + string(b) + ABCI_REQ
	return s, nil
}

// GetRequestFromString parse string and try to get a string of a Request created by GetRequestString.
func GetRequestFromString(s string) (*abci.Request, error) {
	if !strings.Contains(s, ABCI_REQ) {
		return nil, fmt.Errorf("String passed to GetRequestFromString does not have any abci request!\n")
	}
	parts := strings.Split(s, ABCI_REQ)
	if len(parts) != 3 {
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
