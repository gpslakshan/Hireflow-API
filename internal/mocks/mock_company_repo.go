package mocks

import (
	"github.com/google/uuid"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

type MockCompanyRepo struct {
	mock.Mock
}

func (m *MockCompanyRepo) Create(company *entity.Company) error {
	args := m.Called(company)
	return args.Error(0)
}

func (m *MockCompanyRepo) FindAll() ([]entity.Company, error) {
	args := m.Called()
	return args.Get(0).([]entity.Company), args.Error(1)
}

func (m *MockCompanyRepo) FindByID(id uuid.UUID) (*entity.Company, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Company), args.Error(1)
}

func (m *MockCompanyRepo) Update(company *entity.Company) error {
	args := m.Called(company)
	return args.Error(0)
}

func (m *MockCompanyRepo) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}
