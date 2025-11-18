package auth

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,min=6,max=32"`
	Password string `json:"password" validate:"required,min=8,max=30"`
}

type LoginResponse struct {
	Email         string `json:"email"`
	Name          string `json:"name"`
	StampsCounter int    `json:"stampsCounter"`
	Token         string `json:"token"`
	RefreshToken  string `json:"refreshToken"`
}

type RecoveryPasswordRequest struct {
	Email string `json:"email"`
}

type RecoveryPasswordRequestResponse struct {
	Email string `json:"email"`
	Token string `json:"token"`
}
type UnlockUserReq struct {
	UserId uint `json:"userId"`
}
