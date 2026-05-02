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

// Create godoc
// @Summary      Create a company
// @Description  Creates a new company. Admin only.
// @Tags         Companies
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.CreateCompanyRequest  true  "Company details"
// @Success      201      {object}  handler.APIResponse{data=dto.CompanyResponse}
// @Failure      400      {object}  handler.APIResponse
// @Failure      401      {object}  handler.APIResponse
// @Failure      403      {object}  handler.APIResponse
// @Failure      422      {object}  handler.APIResponse{errors=map[string]string}
// @Router       /companies [post]
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

// GetAll godoc
// @Summary      List all companies
// @Description  Returns a list of all companies. Public endpoint.
// @Tags         Companies
// @Produce      json
// @Success      200  {object}  handler.APIResponse{data=[]dto.CompanyResponse}
// @Failure      500  {object}  handler.APIResponse
// @Router       /companies [get]
func (h *CompanyHandler) GetAll(c *gin.Context) {
	companies, err := h.companyService.GetAll()
	if err != nil {
		InternalServerError(c, "failed to fetch companies")
		return
	}

	OK(c, "companies retrieved successfully", mapper.ToCompanyResponseList(companies))
}

// GetByID godoc
// @Summary      Get a company
// @Description  Returns a single company by ID. Public endpoint.
// @Tags         Companies
// @Produce      json
// @Param        id   path      string  true  "Company UUID"
// @Success      200  {object}  handler.APIResponse{data=dto.CompanyResponse}
// @Failure      400  {object}  handler.APIResponse
// @Failure      404  {object}  handler.APIResponse
// @Router       /companies/{id} [get]
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

// Update godoc
// @Summary      Update a company
// @Description  Updates company details. Recruiter only.
// @Tags         Companies
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                    true  "Company UUID"
// @Param        request  body      dto.UpdateCompanyRequest  true  "Fields to update"
// @Success      200      {object}  handler.APIResponse{data=dto.CompanyResponse}
// @Failure      400      {object}  handler.APIResponse
// @Failure      401      {object}  handler.APIResponse
// @Failure      403      {object}  handler.APIResponse
// @Failure      404      {object}  handler.APIResponse
// @Router       /companies/{id} [put]
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

// Delete godoc
// @Summary      Delete a company
// @Description  Soft-deletes a company. Admin only.
// @Tags         Companies
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Company UUID"
// @Success      204  "No Content"
// @Failure      400  {object}  handler.APIResponse
// @Failure      401  {object}  handler.APIResponse
// @Failure      403  {object}  handler.APIResponse
// @Failure      404  {object}  handler.APIResponse
// @Router       /companies/{id} [delete]
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
