package dto

import "cloud.google.com/go/civil"

type CreateSessionResp struct {
	AccessToken  string
	RefreshToken string
}

type RenewTokenReq struct {
	AccessToken  string
	RefreshToken string
}

type RenewTokenResp struct {
	AccessToken  string
	RefreshToken string
}

type Session struct {
	ID                 string
	UserID             string
	HashedRefreshToken string
	IsRevoked          bool
	UserAgent          string
	IPAddress          string
	RevokedAt          civil.Time
	ExpiresAt          civil.Time
	CreatedAt          civil.Time
}
