package dto

const (
	UserRoleAdmin = "admin"
	UserRoleUser  = "user"
)

type HealthCheckResp struct {
	Status string `json:"status"`
}

type CreateUserReq struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type CreateUserResp struct {
	UserID       string `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"-"` // Only set in HttpOnly cookie, never in JSON body
}
