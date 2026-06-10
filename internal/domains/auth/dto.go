package auth

type RegisterDTO struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type LoginDTO struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	SessionSecret string `json:"session_secret"`
	SessionID string `json:"session_id"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}