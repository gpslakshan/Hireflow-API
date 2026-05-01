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

// ── helpers ────────────────────────────────────────────────────────────────

func newAppService(appRepo *mocks.MockAppRepo, jobRepo *mocks.MockJobRepo) *service.ApplicationService {
	return service.NewApplicationService(appRepo, jobRepo)
}

// ── Apply ──────────────────────────────────────────────────────────────────

// Happy path — application created with applied status
func TestApply_Success(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	svc := newAppService(appRepo, jobRepo)

	jobID := uuid.New()
	candidateID := uuid.New()

	openJob := &entity.Job{
		ID:     jobID,
		Status: entity.JobStatusOpen,
	}

	// Tell the mocks what to return
	jobRepo.On("FindByID", jobID).Return(openJob, nil)
	appRepo.On("FindByJobAndCandidate", jobID, candidateID).Return(nil, errors.New("not found"))
	appRepo.On("Create", mock.AnythingOfType("*entity.Application")).Return(nil)
	appRepo.On("FindByID", mock.AnythingOfType("uuid.UUID")).Return(
		&entity.Application{
			ID:          uuid.New(),
			JobID:       jobID,
			CandidateID: candidateID,
			Status:      entity.StatusApplied,
			Job:         *openJob,
		}, nil,
	)

	req := dto.ApplyJobRequest{CoverLetter: "I am a great fit."}
	app, err := svc.Apply(jobID, candidateID, req)

	assert.NoError(t, err)
	assert.Equal(t, entity.StatusApplied, app.Status)
	assert.Equal(t, jobID, app.JobID)
	assert.Equal(t, candidateID, app.CandidateID)

	// Verify all expected mock calls were made
	jobRepo.AssertExpectations(t)
	appRepo.AssertExpectations(t)
}

// Cannot apply to non-existent job
func TestApply_JobNotFound(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	svc := newAppService(appRepo, jobRepo)

	jobID := uuid.New()
	candidateID := uuid.New()

	jobRepo.On("FindByID", jobID).Return(nil, errors.New("not found"))

	_, err := svc.Apply(jobID, candidateID, dto.ApplyJobRequest{})

	assert.ErrorIs(t, err, service.ErrJobNotFound)
	jobRepo.AssertExpectations(t)
	// appRepo should never be called if job doesn't exist
	appRepo.AssertNotCalled(t, "Create")
}

// Cannot apply to a closed job
func TestApply_JobClosed(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	svc := newAppService(appRepo, jobRepo)

	jobID := uuid.New()
	candidateID := uuid.New()

	closedJob := &entity.Job{
		ID:     jobID,
		Status: entity.JobStatusClosed, // ← closed
	}

	jobRepo.On("FindByID", jobID).Return(closedJob, nil)

	_, err := svc.Apply(jobID, candidateID, dto.ApplyJobRequest{})

	assert.ErrorIs(t, err, service.ErrJobClosed)
	appRepo.AssertNotCalled(t, "Create")
}

// Cannot apply to the same job twice
func TestApply_AlreadyApplied(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	svc := newAppService(appRepo, jobRepo)

	jobID := uuid.New()
	candidateID := uuid.New()

	openJob := &entity.Job{ID: jobID, Status: entity.JobStatusOpen}
	existingApp := &entity.Application{ID: uuid.New(), JobID: jobID, CandidateID: candidateID}

	jobRepo.On("FindByID", jobID).Return(openJob, nil)
	// Simulate existing application found
	appRepo.On("FindByJobAndCandidate", jobID, candidateID).Return(existingApp, nil)

	_, err := svc.Apply(jobID, candidateID, dto.ApplyJobRequest{})

	assert.ErrorIs(t, err, service.ErrAlreadyApplied)
	appRepo.AssertNotCalled(t, "Create")
}

// ── UpdateStatus ───────────────────────────────────────────────────────────

// Recruiter can advance pipeline stage
func TestUpdateStatus_Success(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	svc := newAppService(appRepo, jobRepo)

	recruiterID := uuid.New()
	appID := uuid.New()

	existingApp := &entity.Application{
		ID:     appID,
		Status: entity.StatusApplied,
		Job: entity.Job{
			PostedBy: recruiterID, // same recruiter
		},
	}

	appRepo.On("FindByID", appID).Return(existingApp, nil)
	appRepo.On("Update", mock.AnythingOfType("*entity.Application")).Return(nil)

	req := dto.UpdateApplicationStatusRequest{Status: "screening"}
	updated, err := svc.UpdateStatus(appID, recruiterID, req)

	assert.NoError(t, err)
	assert.Equal(t, entity.ApplicationStatus("screening"), updated.Status)
	appRepo.AssertExpectations(t)
}

// Only the job's poster can update status
func TestUpdateStatus_UnauthorizedRecruiter(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	svc := newAppService(appRepo, jobRepo)

	recruiterID := uuid.New()
	differentRecruiterID := uuid.New() // ← different recruiter
	appID := uuid.New()

	existingApp := &entity.Application{
		ID: appID,
		Job: entity.Job{
			PostedBy: differentRecruiterID, // job owned by someone else
		},
	}

	appRepo.On("FindByID", appID).Return(existingApp, nil)

	req := dto.UpdateApplicationStatusRequest{Status: "screening"}
	_, err := svc.UpdateStatus(appID, recruiterID, req)

	assert.ErrorIs(t, err, service.ErrUnauthorizedApplication)
	appRepo.AssertNotCalled(t, "Update")
}

// Returns correct error for missing application
func TestUpdateStatus_ApplicationNotFound(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	svc := newAppService(appRepo, jobRepo)

	appRepo.On("FindByID", mock.AnythingOfType("uuid.UUID")).
		Return(nil, errors.New("not found"))

	_, err := svc.UpdateStatus(uuid.New(), uuid.New(), dto.UpdateApplicationStatusRequest{})

	assert.ErrorIs(t, err, service.ErrApplicationNotFound)
}

// ── Withdraw ───────────────────────────────────────────────────────────────

// Candidate can withdraw their own application
func TestWithdraw_Success(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	svc := newAppService(appRepo, jobRepo)

	candidateID := uuid.New()
	appID := uuid.New()

	existingApp := &entity.Application{
		ID:          appID,
		CandidateID: candidateID, // same candidate
	}

	appRepo.On("FindByID", appID).Return(existingApp, nil)
	appRepo.On("Delete", appID).Return(nil)

	err := svc.Withdraw(appID, candidateID)

	assert.NoError(t, err)
	appRepo.AssertExpectations(t)
}

// Cannot withdraw someone else's application
func TestWithdraw_NotOwner(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	svc := newAppService(appRepo, jobRepo)

	candidateID := uuid.New()
	differentCandidateID := uuid.New()
	appID := uuid.New()

	existingApp := &entity.Application{
		ID:          appID,
		CandidateID: differentCandidateID, // owned by someone else
	}

	appRepo.On("FindByID", appID).Return(existingApp, nil)

	err := svc.Withdraw(appID, candidateID)

	assert.ErrorIs(t, err, service.ErrUnauthorizedApplication)
	appRepo.AssertNotCalled(t, "Delete")
}
