package update_user

type UpdateUserDto struct {
	FirstName string `json:"first_name" validate:"required,min=3,max=64"`
	LastName  string `json:"last_name" validate:"required,min=3,max=64"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"required,numeric"`
	Role      string `json:"role" validate:"required"`
}
