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
	ErrJobNotFound      = errors.New("job not found")
	ErrJobAlreadyClosed = errors.New("job is already closed")
	ErrUnauthorizedJob  = errors.New("you are not authorized to manage this job")
)

type JobService struct {
	jobRepo     repository.JobRepo
	companyRepo repository.CompanyRepo
}

func NewJobService(
	jobRepo repository.JobRepo,
	companyRepo repository.CompanyRepo,
) *JobService {
	return &JobService{jobRepo: jobRepo, companyRepo: companyRepo}
}

func (s *JobService) Create(companyID uuid.UUID, posterID uuid.UUID, req dto.CreateJobRequest) (entity.Job, error) {
	// Verify the company exists
	_, err := s.companyRepo.FindByID(companyID)
	if err != nil {
		return entity.Job{}, ErrCompanyNotFound
	}

	job := entity.Job{
		CompanyID:   companyID,
		PostedBy:    posterID,
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		JobType:     entity.JobType(req.JobType),
		Status:      entity.JobStatusOpen,
	}

	if err := s.jobRepo.Create(&job); err != nil {
		return entity.Job{}, fmt.Errorf("failed to create job: %w", err)
	}

	// After saving, the 'job' struct is incomplete (missing related data like the company).
	// We re-fetch the job from the DB using Preload("Company") to ensure
	// the returned object is fully populated for the frontend.
	created, err := s.jobRepo.FindByID(job.ID)
	if err != nil {
		return entity.Job{}, fmt.Errorf("failed to reload job: %w", err)
	}

	return *created, nil
}

func (s *JobService) GetAll() ([]entity.Job, error) {
	return s.jobRepo.FindAll()
}

func (s *JobService) GetByID(id uuid.UUID) (entity.Job, error) {
	job, err := s.jobRepo.FindByID(id)
	if err != nil {
		return entity.Job{}, ErrJobNotFound
	}
	return *job, nil
}

func (s *JobService) Update(id uuid.UUID, recruiterID uuid.UUID, req dto.UpdateJobRequest) (entity.Job, error) {
	job, err := s.jobRepo.FindByID(id)
	if err != nil {
		return entity.Job{}, ErrJobNotFound
	}

	// ✅ Business rule: only the recruiter who posted this job can update it
	if job.PostedBy != recruiterID {
		return entity.Job{}, ErrUnauthorizedJob
	}

	if req.Title != "" {
		job.Title = req.Title
	}
	if req.Description != "" {
		job.Description = req.Description
	}
	if req.Location != "" {
		job.Location = req.Location
	}
	if req.JobType != "" {
		job.JobType = entity.JobType(req.JobType)
	}

	if err := s.jobRepo.Update(job); err != nil {
		return entity.Job{}, fmt.Errorf("failed to update job: %w", err)
	}

	return *job, nil
}

func (s *JobService) Close(id uuid.UUID, recruiterID uuid.UUID) (entity.Job, error) {
	job, err := s.jobRepo.FindByID(id)
	if err != nil {
		return entity.Job{}, ErrJobNotFound
	}

	// Business rule: only the poster can close a job
	if job.PostedBy != recruiterID {
		return entity.Job{}, ErrUnauthorizedJob
	}

	// Business rule: closing an already closed job is a no-op error
	if job.Status == entity.JobStatusClosed {
		return entity.Job{}, ErrJobAlreadyClosed
	}

	job.Status = entity.JobStatusClosed

	if err := s.jobRepo.Update(job); err != nil {
		return entity.Job{}, fmt.Errorf("failed to close job: %w", err)
	}

	return *job, nil
}

func (s *JobService) Delete(id uuid.UUID, recruiterID uuid.UUID) error {
	job, err := s.jobRepo.FindByID(id)
	if err != nil {
		return ErrJobNotFound
	}

	// Business rule: only the poster can delete a job
	if job.PostedBy != recruiterID {
		return ErrUnauthorizedJob
	}

	return s.jobRepo.Delete(id)
}
