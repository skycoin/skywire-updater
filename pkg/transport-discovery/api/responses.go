package api

// NonceResponse contains nonce related endpoints response values.
type NonceResponse struct {
	Edge      string `json:"edge"`
	NextNonce uint64 `json:"next_nonce"`
}