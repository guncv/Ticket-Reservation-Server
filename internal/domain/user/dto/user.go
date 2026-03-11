package dto

type HealthCheckResp struct {
	Status string `json:"status"`
}

type GetOAuthURLResult struct {
	AuthURL string `json:"auth_url"`
}

type SignInOAuthReq struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type SignInResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
