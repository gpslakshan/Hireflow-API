package mocks

import (
	"github.com/google/uuid"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

// MockAppRepo is a testify mock that satisfies repository.AppRepo
type MockAppRepo struct {
	mock.Mock
}

func (m *MockAppRepo) Create(app *entity.Application) error {
	args := m.Called(app)
	return args.Error(0)
}

func (m *MockAppRepo) FindByID(id uuid.UUID) (*entity.Application, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Application), args.Error(1)
}

func (m *MockAppRepo) FindByCandidate(candidateID uuid.UUID) ([]entity.Application, error) {
	args := m.Called(candidateID)
	return args.Get(0).([]entity.Application), args.Error(1)
}

func (m *MockAppRepo) FindByJob(jobID uuid.UUID) ([]entity.Application, error) {
	args := m.Called(jobID)
	return args.Get(0).([]entity.Application), args.Error(1)
}

func (m *MockAppRepo) FindByJobAndCandidate(jobID, candidateID uuid.UUID) (*entity.Application, error) {
	args := m.Called(jobID, candidateID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Application), args.Error(1)
}

func (m *MockAppRepo) Update(app *entity.Application) error {
	args := m.Called(app)
	return args.Error(0)
}

func (m *MockAppRepo) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}
