package app

import (
	"encoding/base64"
	"fmt"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
)

const ABCI_REQ = "abci-req"

// GetRequestString gets the string representation of the request that will be logged by the application.
func GetABCIRequestString(req *abci.Request) (string, error) {
	b, err := proto.Marshal(req)
	if err != nil {
		return "", err
	}
	reqStr := base64.StdEncoding.EncodeToString(b)
	s := ABCI_REQ + reqStr + ABCI_REQ
	return s, nil
}

// GetABCIRequestFromString parse string and try to get a string of a Request created by GetRequestString.
func GetABCIRequestFromString(s string) (*abci.Request, error) {
	if !strings.Contains(s, ABCI_REQ) {
		return nil, nil
	}
	parts := strings.Split(s, ABCI_REQ)
	if len(parts) != 3 {
		return nil, fmt.Errorf("String %v passed to GetRequestFromString does not have a good format!\n", s)
	}
	req := &abci.Request{}
	reqStr := parts[1]
	b, err := base64.StdEncoding.DecodeString(reqStr)
	if err != nil {
		return nil, fmt.Errorf("String %v cannot be decoded to bytes.\n Error: %v\n ", s, err.Error())
	}
	err = proto.Unmarshal(b, req)
	if err != nil {
		return nil, err
	}
	return req, nil
}
