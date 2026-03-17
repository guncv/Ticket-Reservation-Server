package dto

type UserRole = string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

type CreateUserReq struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type CreateUserResp struct {
	UserID       string `json:"user_id"`
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

type User struct {
	ID             string
	UserName       string
	HashedPassword string
	Role           string
}
