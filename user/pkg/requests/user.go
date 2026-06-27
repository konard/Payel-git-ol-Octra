package requests

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type SubscribeRequest struct {
	UserID string `json:"user_id"`
	Plan   string `json:"plan"`
}

type PromoCodeRequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

type CreatePaymentRequest struct {
	PlanID    string `json:"plan_id" binding:"required"`
	ReturnURL string `json:"return_url" binding:"required"`
}

type SimulatePaymentRequest struct {
	UserID string `json:"user_id" binding:"required"`
	PlanID string `json:"plan_id" binding:"required"`
}
