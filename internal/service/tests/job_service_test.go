package tests

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/gpslakshan/hireflow/internal/domain/dto"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
	"github.com/gpslakshan/hireflow/internal/mocks"
	"github.com/gpslakshan/hireflow/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newJobService(jobRepo *mocks.MockJobRepo, companyRepo *mocks.MockCompanyRepo) *service.JobService {
	return service.NewJobService(jobRepo, companyRepo)
}

// ── Close ──────────────────────────────────────────────────────────────────

// Recruiter can close their open job
func TestClose_Success(t *testing.T) {
	jobRepo := new(mocks.MockJobRepo)
	companyRepo := new(mocks.MockCompanyRepo)
	svc := newJobService(jobRepo, companyRepo)

	recruiterID := uuid.New()
	jobID := uuid.New()

	openJob := &entity.Job{
		ID:       jobID,
		PostedBy: recruiterID,
		Status:   entity.JobStatusOpen,
		Company:  entity.Company{},
	}

	jobRepo.On("FindByID", jobID).Return(openJob, nil)
	jobRepo.On("Update", mock.AnythingOfType("*entity.Job")).Return(nil)

	updated, err := svc.Close(jobID, recruiterID)

	assert.NoError(t, err)
	assert.Equal(t, entity.JobStatusClosed, updated.Status)
	jobRepo.AssertExpectations(t)
}

// Cannot close an already closed job
func TestClose_AlreadyClosed(t *testing.T) {
	jobRepo := new(mocks.MockJobRepo)
	companyRepo := new(mocks.MockCompanyRepo)
	svc := newJobService(jobRepo, companyRepo)

	recruiterID := uuid.New()
	jobID := uuid.New()

	closedJob := &entity.Job{
		ID:       jobID,
		PostedBy: recruiterID,
		Status:   entity.JobStatusClosed, // ← already closed
	}

	jobRepo.On("FindByID", jobID).Return(closedJob, nil)

	_, err := svc.Close(jobID, recruiterID)

	assert.ErrorIs(t, err, service.ErrJobAlreadyClosed)
	jobRepo.AssertNotCalled(t, "Update")
}

// Only the poster can close a job
func TestClose_UnauthorizedRecruiter(t *testing.T) {
	jobRepo := new(mocks.MockJobRepo)
	companyRepo := new(mocks.MockCompanyRepo)
	svc := newJobService(jobRepo, companyRepo)

	recruiterID := uuid.New()
	differentRecruiterID := uuid.New()
	jobID := uuid.New()

	job := &entity.Job{
		ID:       jobID,
		PostedBy: differentRecruiterID, // ← someone else's job
		Status:   entity.JobStatusOpen,
	}

	jobRepo.On("FindByID", jobID).Return(job, nil)

	_, err := svc.Close(jobID, recruiterID)

	assert.ErrorIs(t, err, service.ErrUnauthorizedJob)
	jobRepo.AssertNotCalled(t, "Update")
}

// Returns correct error for missing job
func TestClose_JobNotFound(t *testing.T) {
	jobRepo := new(mocks.MockJobRepo)
	companyRepo := new(mocks.MockCompanyRepo)
	svc := newJobService(jobRepo, companyRepo)

	jobRepo.On("FindByID", mock.AnythingOfType("uuid.UUID")).
		Return(nil, errors.New("not found"))

	_, err := svc.Close(uuid.New(), uuid.New())

	assert.ErrorIs(t, err, service.ErrJobNotFound)
}

// ── Delete ─────────────────────────────────────────────────────────────────

// Recruiter can delete their own job
func TestDeleteJob_Success(t *testing.T) {
	jobRepo := new(mocks.MockJobRepo)
	companyRepo := new(mocks.MockCompanyRepo)
	svc := newJobService(jobRepo, companyRepo)

	recruiterID := uuid.New()
	jobID := uuid.New()

	job := &entity.Job{
		ID:       jobID,
		PostedBy: recruiterID,
	}

	jobRepo.On("FindByID", jobID).Return(job, nil)
	jobRepo.On("Delete", jobID).Return(nil)

	err := svc.Delete(jobID, recruiterID)

	assert.NoError(t, err)
	jobRepo.AssertExpectations(t)
}

// Cannot delete someone else's job
func TestDeleteJob_Unauthorized(t *testing.T) {
	jobRepo := new(mocks.MockJobRepo)
	companyRepo := new(mocks.MockCompanyRepo)
	svc := newJobService(jobRepo, companyRepo)

	recruiterID := uuid.New()
	differentRecruiterID := uuid.New()
	jobID := uuid.New()

	job := &entity.Job{
		ID:       jobID,
		PostedBy: differentRecruiterID,
	}

	jobRepo.On("FindByID", jobID).Return(job, nil)

	err := svc.Delete(jobID, recruiterID)

	assert.ErrorIs(t, err, service.ErrUnauthorizedJob)
	jobRepo.AssertNotCalled(t, "Delete")
}

// ── Update ─────────────────────────────────────────────────────────────────

// Job fields update correctly
func TestUpdateJob_Success(t *testing.T) {
	jobRepo := new(mocks.MockJobRepo)
	companyRepo := new(mocks.MockCompanyRepo)
	svc := newJobService(jobRepo, companyRepo)

	recruiterID := uuid.New()
	jobID := uuid.New()

	job := &entity.Job{
		ID:       jobID,
		PostedBy: recruiterID,
		Title:    "Old Title",
		Company:  entity.Company{},
	}

	jobRepo.On("FindByID", jobID).Return(job, nil)
	jobRepo.On("Update", mock.AnythingOfType("*entity.Job")).Return(nil)

	req := dto.UpdateJobRequest{Title: "New Title"}
	updated, err := svc.Update(jobID, recruiterID, req)

	assert.NoError(t, err)
	assert.Equal(t, "New Title", updated.Title)
	jobRepo.AssertExpectations(t)
}
