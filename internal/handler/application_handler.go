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
