package repository

import (
	"github.com/google/uuid"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
)

type UserRepo interface {
	Create(user *entity.User) error
	FindByEmail(email string) (*entity.User, error)
	FindByID(id string) (*entity.User, error)
}

type CompanyRepo interface {
	Create(company *entity.Company) error
	FindAll() ([]entity.Company, error)
	FindByID(id uuid.UUID) (*entity.Company, error)
	Update(company *entity.Company) error
	Delete(id uuid.UUID) error
}

type JobRepo interface {
	Create(job *entity.Job) error
	FindAll() ([]entity.Job, error)
	FindByID(id uuid.UUID) (*entity.Job, error)
	FindByCompanyID(companyID uuid.UUID) ([]entity.Job, error)
	Update(job *entity.Job) error
	Delete(id uuid.UUID) error
}

type AppRepo interface {
	Create(app *entity.Application) error
	FindByID(id uuid.UUID) (*entity.Application, error)
	FindByCandidate(candidateID uuid.UUID) ([]entity.Application, error)
	FindByJob(jobID uuid.UUID) ([]entity.Application, error)
	FindByJobAndCandidate(jobID, candidateID uuid.UUID) (*entity.Application, error)
	Update(app *entity.Application) error
	Delete(id uuid.UUID) error
}
