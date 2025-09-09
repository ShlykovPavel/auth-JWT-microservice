package get_user_by_id

type UserInfo struct {
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"required,numeric"`
	LastName  string `json:"last_name"`
	FirstName string `json:"first_name"`
}
