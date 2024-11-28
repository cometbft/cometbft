package internal

type NewBlock struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      string `json:"id"`
	Result  struct {
		Query string `json:"query"`
		Data  struct {
			Type  string `json:"type"`
			Value struct {
				Block struct {
					Header struct {
						Height string `json:"height"`
						Time   string `json:"time"`
					} `json:"header"`
					Data struct {
						Txs []string `json:"txs"`
					} `json:"data"`
				} `json:"block"`
			} `json:"value"`
		} `json:"data"`
	} `json:"result"`
}
