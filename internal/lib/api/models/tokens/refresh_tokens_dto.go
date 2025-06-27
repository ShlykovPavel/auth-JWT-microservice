package tokens

type RefreshTokensDto struct {
	AccessToken  string `json:"access_token" validate:"required,min=3"`
	RefreshToken string `json:"refresh_token" validate:"required,min=3"`
}
