package biz

type RegisterRequest struct {
	Email    string
	Password string
}

type RegisterResponse struct {
	UserID string
}

type LoginRequest struct {
	Email    string
	Password string
}

type LoginResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	UserID       string
}

type RefreshTokenRequest struct {
	RefreshToken string
}

type RefreshTokenResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	UserID       string
}

type LogoutRequest struct {
	RefreshToken string
}

type LogoutResponse struct {
	Success bool
}
