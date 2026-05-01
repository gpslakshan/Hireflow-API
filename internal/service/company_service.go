package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gpslakshan/hireflow/internal/domain/dto"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
	"github.com/gpslakshan/hireflow/internal/repository"
)

var (
	ErrCompanyNotFound  = errors.New("company not found")
	ErrCompanyNameTaken = errors.New("company name already exists")
)

type CompanyService struct {
	companyRepo repository.CompanyRepo
}

func NewCompanyService(companyRepo repository.CompanyRepo) *CompanyService {
	return &CompanyService{companyRepo: companyRepo}
}

func (s *CompanyService) Create(req dto.CreateCompanyRequest) (entity.Company, error) {
	company := entity.Company{
		Name:        req.Name,
		Description: req.Description,
		Industry:    req.Industry,
		Website:     req.Website,
		Location:    req.Location,
	}

	if err := s.companyRepo.Create(&company); err != nil {
		return entity.Company{}, fmt.Errorf("failed to create company: %w", err)
	}

	return company, nil
}

func (s *CompanyService) GetAll() ([]entity.Company, error) {
	return s.companyRepo.FindAll()
}

func (s *CompanyService) GetByID(id uuid.UUID) (entity.Company, error) {
	company, err := s.companyRepo.FindByID(id)
	if err != nil {
		return entity.Company{}, ErrCompanyNotFound
	}
	return *company, nil
}

func (s *CompanyService) Update(id uuid.UUID, req dto.UpdateCompanyRequest) (entity.Company, error) {
	// 1. Fetch existing — fail fast if not found
	company, err := s.companyRepo.FindByID(id)
	if err != nil {
		return entity.Company{}, ErrCompanyNotFound
	}

	// 2. Only update fields that were actually provided
	if req.Name != "" {
		company.Name = req.Name
	}
	if req.Description != "" {
		company.Description = req.Description
	}
	if req.Industry != "" {
		company.Industry = req.Industry
	}
	if req.Website != "" {
		company.Website = req.Website
	}
	if req.Location != "" {
		company.Location = req.Location
	}

	if err := s.companyRepo.Update(company); err != nil {
		return entity.Company{}, fmt.Errorf("failed to update company: %w", err)
	}

	return *company, nil
}

func (s *CompanyService) Delete(id uuid.UUID) error {
	_, err := s.companyRepo.FindByID(id)
	if err != nil {
		return ErrCompanyNotFound
	}
	return s.companyRepo.Delete(id)
}
