package biz

type M2MAuthRequest struct {
	ClientID     string
	ClientSecret string
}

type M2MAuthResponse struct {
	AccessToken string
	ExpiresIn   int64
}
