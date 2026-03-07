package users_outbox_db

import (
	"github.com/stretchr/testify/mock"
)

type MockUsersOutboxRepository struct {
	mock.Mock
}

func (m *MockUsersOutboxRepository) GetUnsentUsers() ([]User, error) {
	args := m.Called()
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockUsersOutboxRepository) UpdateAttemptCount(usersId []int64) (int64, error) {
	args := m.Called(usersId)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUsersOutboxRepository) MarkAsSentToKafka(usersId []int64) (int64, error) {
	args := m.Called(usersId)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUsersOutboxRepository) AddUserToOutbox(userId int64, eventType string) error {
	args := m.Called(userId, eventType)
	return args.Error(0)
}

func (m *MockUsersOutboxRepository) GetAttemptCountLimitList(attemptLimit int8) ([]int64, error) {
	args := m.Called(attemptLimit)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockUsersOutboxRepository) MarkAsFailed(usersId []int64) error {
	args := m.Called(usersId)
	return args.Error(0)
}
