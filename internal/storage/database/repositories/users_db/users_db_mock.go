package users_db

import (
	"context"

	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/query_params"
	"github.com/ShlykovPavel/auth-JWT-microservice/models/users/create_user"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, userinfo *create_user.UserCreate) (int64, error) {
	args := m.Called(ctx, userinfo)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, userEmail string) (UserInfo, error) {
	args := m.Called(ctx, userEmail)
	return args.Get(0).(UserInfo), args.Error(1)
}

func (m *MockUserRepository) GetUser(ctx context.Context, userId int64) (UserInfo, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).(UserInfo), args.Error(1)
}

func (m *MockUserRepository) GetUserList(ctx context.Context, search string, limit, offset int, sortParams []query_params.SortParam) (UserListResult, error) {
	args := m.Called(ctx, search, limit, offset, sortParams)
	return args.Get(0).(UserListResult), args.Error(1)
}

func (m *MockUserRepository) CheckAdminInDB(ctx context.Context) (UserInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).(UserInfo), args.Error(1)
}

func (m *MockUserRepository) AddFirstAdmin(ctx context.Context, passwordHash string) error {
	args := m.Called(ctx, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) SetAdminRole(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, id int64, firstName, lastName, email, phone, role string) error {
	args := m.Called(ctx, id, firstName, lastName, email, phone, role)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
