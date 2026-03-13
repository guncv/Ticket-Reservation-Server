package dto

import "time"

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
	RevokedAt          *time.Time
	ExpiresAt          time.Time
	CreatedAt          time.Time
}
