package create_user

type UserCreate struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password"  validate:"required,min=3,max=64"`
}
