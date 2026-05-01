package mocks

import (
	"github.com/google/uuid"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

// MockJobRepo is a testify mock that satisfies repository.JobRepo
type MockJobRepo struct {
	mock.Mock
}

func (m *MockJobRepo) Create(job *entity.Job) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockJobRepo) FindAll() ([]entity.Job, error) {
	args := m.Called()
	return args.Get(0).([]entity.Job), args.Error(1)
}

func (m *MockJobRepo) FindByID(id uuid.UUID) (*entity.Job, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Job), args.Error(1)
}

func (m *MockJobRepo) FindByCompanyID(companyID uuid.UUID) ([]entity.Job, error) {
	args := m.Called(companyID)
	return args.Get(0).([]entity.Job), args.Error(1)
}

func (m *MockJobRepo) Update(job *entity.Job) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockJobRepo) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}
