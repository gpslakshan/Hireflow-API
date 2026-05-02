package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/gpslakshan/hireflow/internal/domain/dto"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
)

type AuthServiceInterface interface {
	Register(req dto.RegisterRequest) (entity.User, error)
	Login(req dto.LoginRequest) (string, entity.User, error)
	GetByID(id string) (*entity.User, error)
}

type CompanyServiceInterface interface {
	Create(req dto.CreateCompanyRequest) (entity.Company, error)
	GetAll() ([]entity.Company, error)
	GetByID(id uuid.UUID) (entity.Company, error)
	Update(id uuid.UUID, req dto.UpdateCompanyRequest) (entity.Company, error)
	Delete(id uuid.UUID) error
}

type JobServiceInterface interface {
	Create(companyID uuid.UUID, posterID uuid.UUID, req dto.CreateJobRequest) (entity.Job, error)
	GetAll() ([]entity.Job, error)
	GetByID(id uuid.UUID) (entity.Job, error)
	Update(id uuid.UUID, recruiterID uuid.UUID, req dto.UpdateJobRequest) (entity.Job, error)
	Close(id uuid.UUID, recruiterID uuid.UUID) (entity.Job, error)
	Delete(id uuid.UUID, recruiterID uuid.UUID) error
}

type ApplicationServiceInterface interface {
	Apply(jobID uuid.UUID, candidateID uuid.UUID, req dto.ApplyJobRequest) (entity.Application, error)
	GetMyApplications(candidateID uuid.UUID) ([]entity.Application, []string, error)
	GetByJob(jobID uuid.UUID, recruiterID uuid.UUID) ([]entity.Application, []string, error)
	GetByID(id uuid.UUID, userID uuid.UUID, role string) (entity.Application, string, error)
	UpdateStatus(id uuid.UUID, recruiterID uuid.UUID, req dto.UpdateApplicationStatusRequest) (entity.Application, error)
	Withdraw(id uuid.UUID, candidateID uuid.UUID) error
}

type UploadServiceInterface interface {
	GenerateCVUploadURL(ctx context.Context, candidateID string, fileName string) (dto.CVUploadURLResponse, error)
}

type UserServiceInterface interface {
	AssignCompany(userID string, req dto.AssignCompanyRequest) (entity.User, error)
}
