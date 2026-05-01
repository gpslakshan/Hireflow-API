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
	ErrApplicationNotFound     = errors.New("application not found")
	ErrAlreadyApplied          = errors.New("you have already applied to this job")
	ErrJobClosed               = errors.New("this job is no longer accepting applications")
	ErrUnauthorizedApplication = errors.New("you are not authorized to access this application")
)

type ApplicationService struct {
	appRepo *repository.ApplicationRepository
	jobRepo *repository.JobRepository
}

func NewApplicationService(
	appRepo *repository.ApplicationRepository,
	jobRepo *repository.JobRepository,
) *ApplicationService {
	return &ApplicationService{appRepo: appRepo, jobRepo: jobRepo}
}

func (s *ApplicationService) Apply(jobID uuid.UUID, candidateID uuid.UUID, req dto.ApplyJobRequest) (entity.Application, error) {
	// 1. Check the job exists
	job, err := s.jobRepo.FindByID(jobID)
	if err != nil {
		return entity.Application{}, ErrJobNotFound
	}

	// 2. Business rule: cannot apply to a closed job
	if job.Status == entity.JobStatusClosed {
		return entity.Application{}, ErrJobClosed
	}

	// 3. Business rule: cannot apply to the same job twice
	existing, _ := s.appRepo.FindByJobAndCandidate(jobID, candidateID)
	if existing != nil {
		return entity.Application{}, ErrAlreadyApplied
	}

	// 4. Create the application
	app := entity.Application{
		JobID:       jobID,
		CandidateID: candidateID,
		CoverLetter: req.CoverLetter,
		Status:      entity.StatusApplied,
	}

	if err := s.appRepo.Create(&app); err != nil {
		return entity.Application{}, fmt.Errorf("failed to create application: %w", err)
	}

	// 5. Reload with associations for the response
	created, err := s.appRepo.FindByID(app.ID)
	if err != nil {
		return entity.Application{}, fmt.Errorf("failed to reload application: %w", err)
	}

	return *created, nil
}

func (s *ApplicationService) GetMyApplications(candidateID uuid.UUID) ([]entity.Application, error) {
	return s.appRepo.FindByCandidate(candidateID)
}

func (s *ApplicationService) GetByJob(jobID uuid.UUID, recruiterID uuid.UUID) ([]entity.Application, error) {
	// Business rule: only the recruiter who posted the job can see its applications
	job, err := s.jobRepo.FindByID(jobID)
	if err != nil {
		return nil, ErrJobNotFound
	}

	if job.PostedBy != recruiterID {
		return nil, ErrUnauthorizedJob
	}

	return s.appRepo.FindByJob(jobID)
}

func (s *ApplicationService) GetByID(id uuid.UUID, userID uuid.UUID, role string) (entity.Application, error) {
	app, err := s.appRepo.FindByID(id)
	if err != nil {
		return entity.Application{}, ErrApplicationNotFound
	}

	// Business rule: candidate can only see their own application
	// Recruiter can only see applications for jobs they posted
	if role == string(entity.RoleCandidate) && app.CandidateID != userID {
		return entity.Application{}, ErrUnauthorizedApplication
	}

	if role == string(entity.RoleRecruiter) && app.Job.PostedBy != userID {
		return entity.Application{}, ErrUnauthorizedApplication
	}

	return *app, nil
}

func (s *ApplicationService) UpdateStatus(id uuid.UUID, recruiterID uuid.UUID, req dto.UpdateApplicationStatusRequest) (entity.Application, error) {
	app, err := s.appRepo.FindByID(id)
	if err != nil {
		return entity.Application{}, ErrApplicationNotFound
	}

	// Business rule: only the recruiter who owns the job can update status
	if app.Job.PostedBy != recruiterID {
		return entity.Application{}, ErrUnauthorizedApplication
	}

	app.Status = entity.ApplicationStatus(req.Status)

	if err := s.appRepo.Update(app); err != nil {
		return entity.Application{}, fmt.Errorf("failed to update status: %w", err)
	}

	return *app, nil
}

func (s *ApplicationService) Withdraw(id uuid.UUID, candidateID uuid.UUID) error {
	app, err := s.appRepo.FindByID(id)
	if err != nil {
		return ErrApplicationNotFound
	}

	// Business rule: candidates can only withdraw their own applications
	if app.CandidateID != candidateID {
		return ErrUnauthorizedApplication
	}

	return s.appRepo.Delete(id)
}
