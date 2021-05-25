package account

/// ComerLoginResposne
type ComerLoginResponse struct {
	Nick    string `json:"nick"`
	Avatar  string `json:"avtar"`
	ComerID string `json:"comer_id"`
	Address string `json:"address"`
	Token   string `json:"token"`
	UIN     uint64 `json:"uin"`
}

/// WalletNonceResponse wrap the nonce for formating rule in resposne
type WalletNonceResponse struct {
	Nonce string `json:"nonce"`
}