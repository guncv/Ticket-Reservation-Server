package dto

type UserRole = string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

type HealthCheckResp struct {
	Status string `json:"status"`
}

type CreateUserReq struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type CreateUserResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"-"` // Only set in HttpOnly cookie, never in JSON body
}

type LoginUserReq struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type LoginUserResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"-"` // Only set in HttpOnly cookie, never in JSON body
}
