package dto

type SessionReq struct {
	AccessToken string
	// RefreshToken string
}

type SessionResult struct {
	UserID       string
	AccessToken  string
	RefreshToken string
}
