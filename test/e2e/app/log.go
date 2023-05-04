package app

import (
	"encoding/base64"
	"fmt"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
)

const ABCI_REQ = "abci:"

// GetRequestString gets the string representation of the request that will be logged by the application.
func GetABCIRequestString(req *abci.Request) (string, error) {
	b, err := proto.Marshal(req)
	if err != nil {
		return "", err
	}
	reqStr := base64.StdEncoding.EncodeToString(b)
	s := ABCI_REQ + reqStr
	return s, nil
}

// GetABCIRequestFromString parse string and try to get a string of a Request created by GetRequestString.
func GetABCIRequestFromString(s string) (*abci.Request, error) {
	if !strings.Contains(s, ABCI_REQ) {
		return nil, fmt.Errorf("String passed to GetRequestFromString does not have any abci request!\n")
	}
	parts := strings.Split(s, ABCI_REQ)
	if len(parts) != 2 {
		return nil, fmt.Errorf("String passed to GetRequestFromString does not have a good format!\n")
	}
	req := &abci.Request{}
	reqStr := parts[1]
	b, err := base64.StdEncoding.DecodeString(reqStr)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(b, req)
	if err != nil {
		return nil, err
	}
	return req, nil
}
