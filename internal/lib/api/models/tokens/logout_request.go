package tokens

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,min=3"`
}
