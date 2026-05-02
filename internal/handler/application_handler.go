package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gpslakshan/hireflow/internal/domain/dto"
	"github.com/gpslakshan/hireflow/internal/domain/mapper"
	"github.com/gpslakshan/hireflow/internal/service"
)

type ApplicationHandler struct {
	appService service.ApplicationServiceInterface
	validate   *validator.Validate
}

func NewApplicationHandler(appService service.ApplicationServiceInterface) *ApplicationHandler {
	return &ApplicationHandler{
		appService: appService,
		validate:   validator.New(),
	}
}

// Apply godoc
// @Summary      Apply to a job
// @Description  Submit an application to an open job. Candidate only.
// @Tags         Applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string               true  "Job UUID"
// @Param        request  body      dto.ApplyJobRequest  true  "Cover letter"
// @Success      201      {object}  handler.APIResponse{data=dto.ApplicationResponse}
// @Failure      400      {object}  handler.APIResponse
// @Failure      401      {object}  handler.APIResponse
// @Failure      403      {object}  handler.APIResponse
// @Failure      404      {object}  handler.APIResponse
// @Failure      409      {object}  handler.APIResponse
// @Router       /jobs/{id}/apply [post]
func (h *ApplicationHandler) Apply(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid job id")
		return
	}

	candidateID, err := extractUserID(c)
	if err != nil {
		Unauthorized(c, "unauthorized")
		return
	}

	var req dto.ApplyJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, formatValidationErrors(err))
		return
	}

	app, err := h.appService.Apply(jobID, candidateID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrJobNotFound):
			NotFound(c, err.Error())
		case errors.Is(err, service.ErrJobClosed):
			Conflict(c, err.Error())
		case errors.Is(err, service.ErrAlreadyApplied):
			Conflict(c, err.Error())
		default:
			InternalServerError(c, "failed to submit application")
		}
		return
	}

	Created(c, "application submitted successfully", mapper.ToApplicationResponse(app, ""))
}

// GetMyApplications godoc
// @Summary      My applications
// @Description  Returns all applications submitted by the current candidate.
// @Tags         Applications
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  handler.APIResponse{data=[]dto.ApplicationResponse}
// @Failure      401  {object}  handler.APIResponse
// @Failure      403  {object}  handler.APIResponse
// @Router       /applications/my [get]
func (h *ApplicationHandler) GetMyApplications(c *gin.Context) {
	candidateID, err := extractUserID(c)
	if err != nil {
		Unauthorized(c, "unauthorized")
		return
	}

	apps, urls, err := h.appService.GetMyApplications(candidateID)
	if err != nil {
		InternalServerError(c, "failed to fetch applications")
		return
	}

	OK(c, "applications retrieved successfully", mapper.ToApplicationResponseList(apps, urls))
}

// GetByJob godoc
// @Summary      Applications for a job
// @Description  Lists all applications for a job. Only the recruiter who posted it.
// @Tags         Applications
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Job UUID"
// @Success      200  {object}  handler.APIResponse{data=[]dto.ApplicationResponse}
// @Failure      401  {object}  handler.APIResponse
// @Failure      403  {object}  handler.APIResponse
// @Failure      404  {object}  handler.APIResponse
// @Router       /jobs/{id}/applications [get]
func (h *ApplicationHandler) GetByJob(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid job id")
		return
	}

	recruiterID, err := extractUserID(c)
	if err != nil {
		Unauthorized(c, "unauthorized")
		return
	}

	apps, urls, err := h.appService.GetByJob(jobID, recruiterID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrJobNotFound):
			NotFound(c, err.Error())
		case errors.Is(err, service.ErrUnauthorizedJob):
			Forbidden(c, err.Error())
		default:
			InternalServerError(c, "failed to fetch applications")
		}
		return
	}

	OK(c, "applications retrieved successfully", mapper.ToApplicationResponseList(apps, urls))
}

// GetByID godoc
// @Summary      Get an application
// @Description  Returns a single application. Candidate sees own only; recruiter sees their job's only.
// @Tags         Applications
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Application UUID"
// @Success      200  {object}  handler.APIResponse{data=dto.ApplicationResponse}
// @Failure      400  {object}  handler.APIResponse
// @Failure      401  {object}  handler.APIResponse
// @Failure      403  {object}  handler.APIResponse
// @Failure      404  {object}  handler.APIResponse
// @Router       /applications/{id} [get]
func (h *ApplicationHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid application id")
		return
	}

	userID, err := extractUserID(c)
	if err != nil {
		Unauthorized(c, "unauthorized")
		return
	}

	role, _ := c.Get("role")

	app, downloadURL, err := h.appService.GetByID(id, userID, role.(string))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrApplicationNotFound):
			NotFound(c, err.Error())
		case errors.Is(err, service.ErrUnauthorizedApplication):
			Forbidden(c, err.Error())
		default:
			InternalServerError(c, "failed to fetch application")
		}
		return
	}

	OK(c, "application retrieved successfully", mapper.ToApplicationResponse(app, downloadURL))
}

// UpdateStatus godoc
// @Summary      Update application status
// @Description  Advances or rejects an application in the pipeline. Recruiter only.
// @Tags         Applications
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                                  true  "Application UUID"
// @Param        request  body      dto.UpdateApplicationStatusRequest      true  "New status"
// @Success      200      {object}  handler.APIResponse{data=dto.ApplicationResponse}
// @Failure      400      {object}  handler.APIResponse
// @Failure      401      {object}  handler.APIResponse
// @Failure      403      {object}  handler.APIResponse
// @Failure      404      {object}  handler.APIResponse
// @Router       /applications/{id}/status [patch]
func (h *ApplicationHandler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid application id")
		return
	}

	recruiterID, err := extractUserID(c)
	if err != nil {
		Unauthorized(c, "unauthorized")
		return
	}

	var req dto.UpdateApplicationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, formatValidationErrors(err))
		return
	}

	app, err := h.appService.UpdateStatus(id, recruiterID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrApplicationNotFound):
			NotFound(c, err.Error())
		case errors.Is(err, service.ErrUnauthorizedApplication):
			Forbidden(c, err.Error())
		default:
			InternalServerError(c, "failed to update status")
		}
		return
	}

	OK(c, "application status updated successfully", mapper.ToApplicationResponse(app, ""))
}

// Withdraw godoc
// @Summary      Withdraw an application
// @Description  Permanently removes the candidate's application. Candidate only.
// @Tags         Applications
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Application UUID"
// @Success      204  "No Content"
// @Failure      400  {object}  handler.APIResponse
// @Failure      401  {object}  handler.APIResponse
// @Failure      403  {object}  handler.APIResponse
// @Failure      404  {object}  handler.APIResponse
// @Router       /applications/{id} [delete]
func (h *ApplicationHandler) Withdraw(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid application id")
		return
	}

	candidateID, err := extractUserID(c)
	if err != nil {
		Unauthorized(c, "unauthorized")
		return
	}

	if err := h.appService.Withdraw(id, candidateID); err != nil {
		switch {
		case errors.Is(err, service.ErrApplicationNotFound):
			NotFound(c, err.Error())
		case errors.Is(err, service.ErrUnauthorizedApplication):
			Forbidden(c, err.Error())
		default:
			InternalServerError(c, "failed to withdraw application")
		}
		return
	}

	NoContent(c)
}
