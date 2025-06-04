package client

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/cometbft/cometbft/v2/rpc/jsonrpc/types"
)

const (
	// URIClientRequestID in a request ID used by URIClient.
	URIClientRequestID = types.JSONRPCIntID(-1)
)

// URIClient is a JSON-RPC client, which sends POST form HTTP requests to the
// remote server.
//
// URIClient is safe for concurrent use by multiple goroutines.
type URIClient struct {
	address string
	client  *http.Client
}

var _ HTTPClient = (*URIClient)(nil)

// NewURI returns a new client.
// An error is returned on invalid remote.
// The function panics when remote is nil.
func NewURI(remote string) (*URIClient, error) {
	parsedURL, err := newParsedURL(remote)
	if err != nil {
		return nil, err
	}

	httpClient, err := DefaultHTTPClient(remote)
	if err != nil {
		return nil, err
	}

	parsedURL.SetDefaultSchemeHTTP()

	uriClient := &URIClient{
		address: parsedURL.GetTrimmedURL(),
		client:  httpClient,
	}

	return uriClient, nil
}

// Call issues a POST form HTTP request.
func (c *URIClient) Call(ctx context.Context, method string,
	params map[string]any, result any,
) (any, error) {
	values, err := argsToURLValues(params)
	if err != nil {
		return nil, ErrEncodingParams{Source: err}
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.address+"/"+method,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return nil, ErrCreateRequest{Source: err}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, ErrFailedRequest{Source: err}
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrReadResponse{Source: err}
	}

	return unmarshalResponseBytes(responseBytes, URIClientRequestID, result)
}
