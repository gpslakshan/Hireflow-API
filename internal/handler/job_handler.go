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

func (h *JobHandler) GetAll(c *gin.Context) {
	jobs, err := h.jobService.GetAll()
	if err != nil {
		InternalServerError(c, "failed to fetch jobs")
		return
	}

	OK(c, "jobs retrieved successfully", mapper.ToJobResponseList(jobs))
}

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
