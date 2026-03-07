package user_service

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/query_params"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database/repositories/users_db"
	"github.com/ShlykovPavel/auth-JWT-microservice/models/users/get_user_by_id"
	"github.com/ShlykovPavel/auth-JWT-microservice/models/users/get_users_list"
	"github.com/ShlykovPavel/auth-JWT-microservice/models/users/update_user"
)

func GetUser(log *slog.Logger, userRepository users_db.UserRepository, userId int64, ctx context.Context) (get_user_by_id.UserInfo, error) {
	const op = "internal/lib/services/user_service/user_service.go/GetUser"
	log = log.With(slog.String("op", op),
		slog.String("UserId", strconv.FormatInt(userId, 10)))

	userInfo, err := userRepository.GetUser(ctx, userId)
	if err != nil {
		if errors.Is(err, users_db.ErrUserNotFound) {
			log.Debug("Пользователь не найден", "err", err)
			return get_user_by_id.UserInfo{}, err
		}
		log.Error("Ошибка поиска пользователя в БД", "err", err)
		return get_user_by_id.UserInfo{}, err
	}
	return get_user_by_id.UserInfo{
		Email:     userInfo.Email,
		Phone:     userInfo.Phone,
		LastName:  userInfo.LastName,
		FirstName: userInfo.FirstName,
	}, nil

}

// GetUserList retrieves a list of users from the repository and converts them to DTOs.
// It takes a logger, user repository, and context as input.
// Returns a slice of UserInfoList DTOs or an error if the operation fails.
func GetUserList(log *slog.Logger, userRepository users_db.UserRepository, ctx context.Context, queryParams query_params.ListQueryParams) (get_users_list.UsersList, error) {
	const op = "internal/lib/services/user_service/user_service.go/GetUserList"
	log = log.With(slog.String("op", op))

	result, err := userRepository.GetUserList(ctx, queryParams.Search, queryParams.Limit, queryParams.Offset, queryParams.SortParams)
	if err != nil {
		log.Error("Failed to get users list", "err", err)
		return get_users_list.UsersList{}, err
	}
	userList := make([]get_users_list.UserInfoList, 0, len(result.Users))
	for _, user := range result.Users {
		userInfo := get_users_list.UserInfoList{
			Id:        user.ID,
			Email:     user.Email,
			Phone:     user.Phone,
			LastName:  user.LastName,
			FirstName: user.FirstName,
			Role:      user.Role,
		}
		userList = append(userList, userInfo)
	}
	metaData := get_users_list.UsersListMetaData{
		Page:   queryParams.Page,
		Total:  result.Total,
		Limit:  queryParams.Limit,
		Offset: queryParams.Offset,
	}
	userDto := get_users_list.UsersList{
		Users: userList,
		Meta:  metaData,
	}
	return userDto, nil

}

func UpdateUser(log *slog.Logger, userRepository users_db.UserRepository, ctx context.Context, dto update_user.UpdateUserDto, id int64) error {
	const op = "internal/lib/services/user_service/user_service.go/UpdateUser"
	log = log.With(slog.String("op", op),
		slog.String("UserId", strconv.FormatInt(id, 10)))

	err := userRepository.UpdateUser(ctx, id, dto.FirstName, dto.LastName, dto.Email, dto.Phone, dto.Role)
	if err != nil {
		log.Error("Failed to update user", "err", err)
		return err
	}
	return nil
}

func DeleteUser(log *slog.Logger, userRepository users_db.UserRepository, ctx context.Context, id int64) error {
	const op = "internal/lib/services/user_service/user_service.go/DeleteUser"
	log = log.With(slog.String("op", op),
		slog.String("UserId", strconv.FormatInt(id, 10)))

	err := userRepository.DeleteUser(ctx, id)
	if err != nil {
		log.Error("Failed to delete user", "err", err)
		return err
	}
	return nil
}
