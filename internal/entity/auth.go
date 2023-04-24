package entity

type Token struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        string `json:"expires_in"`
	RefreshExpiresIn string `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
}
