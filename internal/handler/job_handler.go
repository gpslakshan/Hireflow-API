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

type JobHandler struct {
	jobService service.JobServiceInterface
	validate   *validator.Validate
}

func NewJobHandler(jobService service.JobServiceInterface) *JobHandler {
	return &JobHandler{
		jobService: jobService,
		validate:   validator.New(),
	}
}

// Create godoc
// @Summary      Post a job
// @Description  Creates a job under a company. Recruiter only.
// @Tags         Jobs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                 true  "Company UUID"
// @Param        request  body      dto.CreateJobRequest   true  "Job details"
// @Success      201      {object}  handler.APIResponse{data=dto.JobResponse}
// @Failure      400      {object}  handler.APIResponse
// @Failure      401      {object}  handler.APIResponse
// @Failure      403      {object}  handler.APIResponse
// @Failure      404      {object}  handler.APIResponse
// @Failure      422      {object}  handler.APIResponse{errors=map[string]string}
// @Router       /companies/{id}/jobs [post]
func (h *JobHandler) Create(c *gin.Context) {
	companyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid company id")
		return
	}

	recruiterID, err := extractUserID(c)
	if err != nil {
		Unauthorized(c, "unauthorized")
		return
	}

	var req dto.CreateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, formatValidationErrors(err))
		return
	}

	job, err := h.jobService.Create(companyID, recruiterID, req)
	if err != nil {
		if errors.Is(err, service.ErrCompanyNotFound) {
			NotFound(c, err.Error())
			return
		}
		InternalServerError(c, "failed to create job")
		return
	}

	Created(c, "job created successfully", mapper.ToJobResponse(job))
}

// GetAll godoc
// @Summary      List open jobs
// @Description  Returns all open job postings. Public endpoint.
// @Tags         Jobs
// @Produce      json
// @Success      200  {object}  handler.APIResponse{data=[]dto.JobResponse}
// @Failure      500  {object}  handler.APIResponse
// @Router       /jobs [get]
func (h *JobHandler) GetAll(c *gin.Context) {
	jobs, err := h.jobService.GetAll()
	if err != nil {
		InternalServerError(c, "failed to fetch jobs")
		return
	}

	OK(c, "jobs retrieved successfully", mapper.ToJobResponseList(jobs))
}

// GetByID godoc
// @Summary      Get a job
// @Description  Returns a single job posting by ID. Public endpoint.
// @Tags         Jobs
// @Produce      json
// @Param        id   path      string  true  "Job UUID"
// @Success      200  {object}  handler.APIResponse{data=dto.JobResponse}
// @Failure      400  {object}  handler.APIResponse
// @Failure      404  {object}  handler.APIResponse
// @Router       /jobs/{id} [get]
func (h *JobHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid job id")
		return
	}

	job, err := h.jobService.GetByID(id)
	if err != nil {
		if errors.Is(err, service.ErrJobNotFound) {
			NotFound(c, err.Error())
			return
		}
		InternalServerError(c, "failed to fetch job")
		return
	}

	OK(c, "job retrieved successfully", mapper.ToJobResponse(job))
}

// Update godoc
// @Summary      Update a job
// @Description  Updates a job posting. Only the recruiter who posted it.
// @Tags         Jobs
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                true  "Job UUID"
// @Param        request  body      dto.UpdateJobRequest  true  "Fields to update"
// @Success      200      {object}  handler.APIResponse{data=dto.JobResponse}
// @Failure      400      {object}  handler.APIResponse
// @Failure      401      {object}  handler.APIResponse
// @Failure      403      {object}  handler.APIResponse
// @Failure      404      {object}  handler.APIResponse
// @Router       /jobs/{id} [put]
func (h *JobHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid job id")
		return
	}

	recruiterID, err := extractUserID(c)
	if err != nil {
		Unauthorized(c, "unauthorized")
		return
	}

	var req dto.UpdateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, formatValidationErrors(err))
		return
	}

	job, err := h.jobService.Update(id, recruiterID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrJobNotFound):
			NotFound(c, err.Error())
		case errors.Is(err, service.ErrUnauthorizedJob):
			Forbidden(c, err.Error())
		default:
			InternalServerError(c, "failed to update job")
		}
		return
	}

	OK(c, "job updated successfully", mapper.ToJobResponse(job))
}

// Close godoc
// @Summary      Close a job
// @Description  Marks a job as closed. Only the recruiter who posted it.
// @Tags         Jobs
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Job UUID"
// @Success      200  {object}  handler.APIResponse{data=dto.JobResponse}
// @Failure      400  {object}  handler.APIResponse
// @Failure      401  {object}  handler.APIResponse
// @Failure      403  {object}  handler.APIResponse
// @Failure      404  {object}  handler.APIResponse
// @Failure      409  {object}  handler.APIResponse
// @Router       /jobs/{id}/close [patch]
func (h *JobHandler) Close(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid job id")
		return
	}

	recruiterID, err := extractUserID(c)
	if err != nil {
		Unauthorized(c, "unauthorized")
		return
	}

	job, err := h.jobService.Close(id, recruiterID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrJobNotFound):
			NotFound(c, err.Error())
		case errors.Is(err, service.ErrUnauthorizedJob):
			Forbidden(c, err.Error())
		case errors.Is(err, service.ErrJobAlreadyClosed):
			Conflict(c, err.Error())
		default:
			InternalServerError(c, "failed to close job")
		}
		return
	}

	OK(c, "job closed successfully", mapper.ToJobResponse(job))
}

// Delete godoc
// @Summary      Delete a job
// @Description  Deletes a job posting. Only the recruiter who posted it.
// @Tags         Jobs
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Job UUID"
// @Success      204  "No Content"
// @Failure      400  {object}  handler.APIResponse
// @Failure      401  {object}  handler.APIResponse
// @Failure      403  {object}  handler.APIResponse
// @Failure      404  {object}  handler.APIResponse
// @Router       /jobs/{id} [delete]
func (h *JobHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid job id")
		return
	}

	recruiterID, err := extractUserID(c)
	if err != nil {
		Unauthorized(c, "unauthorized")
		return
	}

	if err := h.jobService.Delete(id, recruiterID); err != nil {
		switch {
		case errors.Is(err, service.ErrJobNotFound):
			NotFound(c, err.Error())
		case errors.Is(err, service.ErrUnauthorizedJob):
			Forbidden(c, err.Error())
		default:
			InternalServerError(c, "failed to delete job")
		}
		return
	}

	NoContent(c)
}
