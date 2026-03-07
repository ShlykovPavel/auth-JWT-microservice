package create_user

type UserCreate struct {
	FirstName string `json:"first_name" validate:"required,min=3,max=64"`
	LastName  string `json:"last_name" validate:"required,min=3,max=64"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password"  validate:"required,min=3,max=64"`
	Phone     string `json:"phone" validate:"required,numeric"`
	Username  string `json:"username" validate:"required,min=3,max=64"`
}
