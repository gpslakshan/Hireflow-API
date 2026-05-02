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

func newAppService(
	appRepo *mocks.MockAppRepo,
	jobRepo *mocks.MockJobRepo,
	cvStorage *mocks.MockCVStorage,
) *service.ApplicationService {
	return service.NewApplicationService(appRepo, jobRepo, cvStorage)
}

// ── Apply ──────────────────────────────────────────────────────────────────

func TestApply_Success(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	cvStorage := new(mocks.MockCVStorage)
	svc := newAppService(appRepo, jobRepo, cvStorage)

	jobID := uuid.New()
	candidateID := uuid.New()

	openJob := &entity.Job{
		ID:     jobID,
		Status: entity.JobStatusOpen,
	}

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

	jobRepo.AssertExpectations(t)
	appRepo.AssertExpectations(t)
	// cvStorage should never be called during Apply
	cvStorage.AssertNotCalled(t, "GenerateUploadURL")
	cvStorage.AssertNotCalled(t, "GenerateDownloadURL")
}

func TestApply_WithCVKey(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	cvStorage := new(mocks.MockCVStorage)
	svc := newAppService(appRepo, jobRepo, cvStorage)

	jobID := uuid.New()
	candidateID := uuid.New()
	cvKey := "cvs/da9d5716/alice-smith-cv.pdf"

	openJob := &entity.Job{ID: jobID, Status: entity.JobStatusOpen}

	jobRepo.On("FindByID", jobID).Return(openJob, nil)
	appRepo.On("FindByJobAndCandidate", jobID, candidateID).Return(nil, errors.New("not found"))
	appRepo.On("Create", mock.AnythingOfType("*entity.Application")).Return(nil)
	appRepo.On("FindByID", mock.AnythingOfType("uuid.UUID")).Return(
		&entity.Application{
			ID:          uuid.New(),
			JobID:       jobID,
			CandidateID: candidateID,
			CVKey:       cvKey,
			Status:      entity.StatusApplied,
			Job:         *openJob,
		}, nil,
	)

	req := dto.ApplyJobRequest{
		CoverLetter: "I am a great fit.",
		CVKey:       cvKey,
	}
	app, err := svc.Apply(jobID, candidateID, req)

	assert.NoError(t, err)
	assert.Equal(t, cvKey, app.CVKey)
	// S3 is not called during Apply — only during fetch
	cvStorage.AssertNotCalled(t, "GenerateDownloadURL")
}

func TestApply_JobNotFound(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	cvStorage := new(mocks.MockCVStorage)
	svc := newAppService(appRepo, jobRepo, cvStorage)

	jobID := uuid.New()
	candidateID := uuid.New()

	jobRepo.On("FindByID", jobID).Return(nil, errors.New("not found"))

	_, err := svc.Apply(jobID, candidateID, dto.ApplyJobRequest{})

	assert.ErrorIs(t, err, service.ErrJobNotFound)
	appRepo.AssertNotCalled(t, "Create")
	cvStorage.AssertNotCalled(t, "GenerateDownloadURL")
}

func TestApply_JobClosed(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	cvStorage := new(mocks.MockCVStorage)
	svc := newAppService(appRepo, jobRepo, cvStorage)

	jobID := uuid.New()
	candidateID := uuid.New()

	closedJob := &entity.Job{ID: jobID, Status: entity.JobStatusClosed}

	jobRepo.On("FindByID", jobID).Return(closedJob, nil)

	_, err := svc.Apply(jobID, candidateID, dto.ApplyJobRequest{})

	assert.ErrorIs(t, err, service.ErrJobClosed)
	appRepo.AssertNotCalled(t, "Create")
}

func TestApply_AlreadyApplied(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	cvStorage := new(mocks.MockCVStorage)
	svc := newAppService(appRepo, jobRepo, cvStorage)

	jobID := uuid.New()
	candidateID := uuid.New()

	openJob := &entity.Job{ID: jobID, Status: entity.JobStatusOpen}
	existingApp := &entity.Application{
		ID:          uuid.New(),
		JobID:       jobID,
		CandidateID: candidateID,
	}

	jobRepo.On("FindByID", jobID).Return(openJob, nil)
	appRepo.On("FindByJobAndCandidate", jobID, candidateID).Return(existingApp, nil)

	_, err := svc.Apply(jobID, candidateID, dto.ApplyJobRequest{})

	assert.ErrorIs(t, err, service.ErrAlreadyApplied)
	appRepo.AssertNotCalled(t, "Create")
}

// ── GetByID ────────────────────────────────────────────────────────────────

func TestGetByID_WithCV_GeneratesDownloadURL(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	cvStorage := new(mocks.MockCVStorage)
	svc := newAppService(appRepo, jobRepo, cvStorage)

	recruiterID := uuid.New()
	appID := uuid.New()
	cvKey := "cvs/candidate-id/alice-cv.pdf"
	signedURL := "https://s3.amazonaws.com/hireflow-cvs/cvs/candidate-id/alice-cv.pdf?X-Amz-Signature=abc"

	existingApp := &entity.Application{
		ID:          appID,
		CVKey:       cvKey,
		CandidateID: uuid.New(),
		Job: entity.Job{
			PostedBy: recruiterID,
		},
	}

	appRepo.On("FindByID", appID).Return(existingApp, nil)
	// Expect S3 to be called with the cv_key to generate a download URL
	cvStorage.On("GenerateDownloadURL", mock.Anything, cvKey).Return(signedURL, nil)

	app, downloadURL, err := svc.GetByID(appID, recruiterID, string(entity.RoleRecruiter))

	assert.NoError(t, err)
	assert.Equal(t, cvKey, app.CVKey)
	assert.Equal(t, signedURL, downloadURL)
	cvStorage.AssertExpectations(t)
}

func TestGetByID_WithoutCV_NoDownloadURL(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	cvStorage := new(mocks.MockCVStorage)
	svc := newAppService(appRepo, jobRepo, cvStorage)

	recruiterID := uuid.New()
	appID := uuid.New()

	existingApp := &entity.Application{
		ID:    appID,
		CVKey: "", // no CV uploaded
		Job:   entity.Job{PostedBy: recruiterID},
	}

	appRepo.On("FindByID", appID).Return(existingApp, nil)

	_, downloadURL, err := svc.GetByID(appID, recruiterID, string(entity.RoleRecruiter))

	assert.NoError(t, err)
	assert.Empty(t, downloadURL)
	// S3 must not be called if there is no CV key
	cvStorage.AssertNotCalled(t, "GenerateDownloadURL")
}

func TestGetByID_CandidateCannotSeeOtherApplication(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	cvStorage := new(mocks.MockCVStorage)
	svc := newAppService(appRepo, jobRepo, cvStorage)

	candidateID := uuid.New()
	differentCandidateID := uuid.New()
	appID := uuid.New()

	existingApp := &entity.Application{
		ID:          appID,
		CandidateID: differentCandidateID, // owned by someone else
	}

	appRepo.On("FindByID", appID).Return(existingApp, nil)

	_, _, err := svc.GetByID(appID, candidateID, string(entity.RoleCandidate))

	assert.ErrorIs(t, err, service.ErrUnauthorizedApplication)
	cvStorage.AssertNotCalled(t, "GenerateDownloadURL")
}

// ── UpdateStatus ───────────────────────────────────────────────────────────

func TestUpdateStatus_Success(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	cvStorage := new(mocks.MockCVStorage)
	svc := newAppService(appRepo, jobRepo, cvStorage)

	recruiterID := uuid.New()
	appID := uuid.New()

	existingApp := &entity.Application{
		ID:     appID,
		Status: entity.StatusApplied,
		Job:    entity.Job{PostedBy: recruiterID},
	}

	appRepo.On("FindByID", appID).Return(existingApp, nil)
	appRepo.On("Update", mock.AnythingOfType("*entity.Application")).Return(nil)

	req := dto.UpdateApplicationStatusRequest{Status: "screening"}
	updated, err := svc.UpdateStatus(appID, recruiterID, req)

	assert.NoError(t, err)
	assert.Equal(t, entity.ApplicationStatus("screening"), updated.Status)
	appRepo.AssertExpectations(t)
	cvStorage.AssertNotCalled(t, "GenerateDownloadURL")
}

func TestUpdateStatus_UnauthorizedRecruiter(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	cvStorage := new(mocks.MockCVStorage)
	svc := newAppService(appRepo, jobRepo, cvStorage)

	recruiterID := uuid.New()
	differentRecruiterID := uuid.New()
	appID := uuid.New()

	existingApp := &entity.Application{
		ID:  appID,
		Job: entity.Job{PostedBy: differentRecruiterID},
	}

	appRepo.On("FindByID", appID).Return(existingApp, nil)

	req := dto.UpdateApplicationStatusRequest{Status: "screening"}
	_, err := svc.UpdateStatus(appID, recruiterID, req)

	assert.ErrorIs(t, err, service.ErrUnauthorizedApplication)
	appRepo.AssertNotCalled(t, "Update")
}

func TestUpdateStatus_ApplicationNotFound(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	cvStorage := new(mocks.MockCVStorage)
	svc := newAppService(appRepo, jobRepo, cvStorage)

	appRepo.On("FindByID", mock.AnythingOfType("uuid.UUID")).
		Return(nil, errors.New("not found"))

	_, err := svc.UpdateStatus(uuid.New(), uuid.New(), dto.UpdateApplicationStatusRequest{})

	assert.ErrorIs(t, err, service.ErrApplicationNotFound)
}

// ── Withdraw ───────────────────────────────────────────────────────────────

func TestWithdraw_Success_WithCV(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	cvStorage := new(mocks.MockCVStorage)
	svc := newAppService(appRepo, jobRepo, cvStorage)

	candidateID := uuid.New()
	appID := uuid.New()
	cvKey := "cvs/candidate-id/alice-cv.pdf"

	existingApp := &entity.Application{
		ID:          appID,
		CandidateID: candidateID,
		CVKey:       cvKey, // has a CV
	}

	appRepo.On("FindByID", appID).Return(existingApp, nil)
	// S3 delete must be called since there is a CV key
	cvStorage.On("DeleteObject", mock.Anything, cvKey).Return(nil)
	appRepo.On("Delete", appID).Return(nil)

	err := svc.Withdraw(appID, candidateID)

	assert.NoError(t, err)
	appRepo.AssertExpectations(t)
	// Verify S3 cleanup was called
	cvStorage.AssertExpectations(t)
}

func TestWithdraw_Success_WithoutCV(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	cvStorage := new(mocks.MockCVStorage)
	svc := newAppService(appRepo, jobRepo, cvStorage)

	candidateID := uuid.New()
	appID := uuid.New()

	existingApp := &entity.Application{
		ID:          appID,
		CandidateID: candidateID,
		CVKey:       "", // no CV uploaded
	}

	appRepo.On("FindByID", appID).Return(existingApp, nil)
	appRepo.On("Delete", appID).Return(nil)

	err := svc.Withdraw(appID, candidateID)

	assert.NoError(t, err)
	// S3 must NOT be called if there was no CV
	cvStorage.AssertNotCalled(t, "DeleteObject")
}

func TestWithdraw_NotOwner(t *testing.T) {
	appRepo := new(mocks.MockAppRepo)
	jobRepo := new(mocks.MockJobRepo)
	cvStorage := new(mocks.MockCVStorage)
	svc := newAppService(appRepo, jobRepo, cvStorage)

	candidateID := uuid.New()
	differentCandidateID := uuid.New()
	appID := uuid.New()

	existingApp := &entity.Application{
		ID:          appID,
		CandidateID: differentCandidateID,
	}

	appRepo.On("FindByID", appID).Return(existingApp, nil)

	err := svc.Withdraw(appID, candidateID)

	assert.ErrorIs(t, err, service.ErrUnauthorizedApplication)
	appRepo.AssertNotCalled(t, "Delete")
	cvStorage.AssertNotCalled(t, "DeleteObject")
}
