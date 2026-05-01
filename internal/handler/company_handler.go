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

type CompanyHandler struct {
	companyService service.CompanyServiceInterface
	validate       *validator.Validate
}

func NewCompanyHandler(companyService service.CompanyServiceInterface) *CompanyHandler {
	return &CompanyHandler{
		companyService: companyService,
		validate:       validator.New(),
	}
}

func (h *CompanyHandler) Create(c *gin.Context) {
	var req dto.CreateCompanyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, formatValidationErrors(err))
		return
	}

	company, err := h.companyService.Create(req)
	if err != nil {
		InternalServerError(c, "failed to create company")
		return
	}

	Created(c, "company created successfully", mapper.ToCompanyResponse(company))
}

func (h *CompanyHandler) GetAll(c *gin.Context) {
	companies, err := h.companyService.GetAll()
	if err != nil {
		InternalServerError(c, "failed to fetch companies")
		return
	}

	OK(c, "companies retrieved successfully", mapper.ToCompanyResponseList(companies))
}

func (h *CompanyHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid company id")
		return
	}

	company, err := h.companyService.GetByID(id)
	if err != nil {
		if errors.Is(err, service.ErrCompanyNotFound) {
			NotFound(c, err.Error())
			return
		}
		InternalServerError(c, "failed to fetch company")
		return
	}

	OK(c, "company retrieved successfully", mapper.ToCompanyResponse(company))
}

func (h *CompanyHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid company id")
		return
	}

	var req dto.UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, formatValidationErrors(err))
		return
	}

	company, err := h.companyService.Update(id, req)
	if err != nil {
		if errors.Is(err, service.ErrCompanyNotFound) {
			NotFound(c, err.Error())
			return
		}
		InternalServerError(c, "failed to update company")
		return
	}

	OK(c, "company updated successfully", mapper.ToCompanyResponse(company))
}

func (h *CompanyHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		BadRequest(c, "invalid company id")
		return
	}

	if err := h.companyService.Delete(id); err != nil {
		if errors.Is(err, service.ErrCompanyNotFound) {
			NotFound(c, err.Error())
			return
		}
		InternalServerError(c, "failed to delete company")
		return
	}

	NoContent(c)
}
