package api

type NonceResponse struct {
	Edge      string `json:"edge"`
	NextNonce uint64 `json:"next_nonce"`
}
