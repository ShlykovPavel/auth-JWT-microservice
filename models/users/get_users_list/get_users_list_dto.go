package get_users_list

type UserInfoList struct {
	Id        int64  `json:"id"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"required,numeric"`
	LastName  string `json:"last_name" validate:"required"`
	FirstName string `json:"first_name" validate:"required"`
	Role      string `json:"role" validate:"required"`
}

type UsersListMetaData struct {
	Page   int   `json:"page"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
	Total  int64 `json:"total"`
}

type UsersList struct {
	Users []UserInfoList    `json:"data"`
	Meta  UsersListMetaData `json:"meta"`
}
