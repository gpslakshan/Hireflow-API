package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gpslakshan/hireflow/internal/domain/dto"
	"github.com/gpslakshan/hireflow/internal/domain/mapper"
	"github.com/gpslakshan/hireflow/internal/service"
)

type UserHandler struct {
	userService service.UserServiceInterface
	validate    *validator.Validate
}

func NewUserHandler(userService service.UserServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
		validate:    validator.New(),
	}
}

// AssignCompany godoc
// @Summary      Assign a recruiter to a company
// @Description  Links a recruiter account to a company. Admin only.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                       true  "User UUID"
// @Param        request  body      dto.AssignCompanyRequest     true  "Company to assign"
// @Success      200      {object}  handler.APIResponse{data=dto.UserResponse}
// @Failure      400      {object}  handler.APIResponse
// @Failure      401      {object}  handler.APIResponse
// @Failure      403      {object}  handler.APIResponse
// @Failure      404      {object}  handler.APIResponse
// @Failure      409      {object}  handler.APIResponse
// @Router       /users/{id}/assign-company [patch]
func (h *UserHandler) AssignCompany(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		BadRequest(c, "invalid user id")
		return
	}

	var req dto.AssignCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, formatValidationErrors(err))
		return
	}

	user, err := h.userService.AssignCompany(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			NotFound(c, err.Error())
		case errors.Is(err, service.ErrNotARecruiter):
			BadRequest(c, err.Error())
		case errors.Is(err, service.ErrAlreadyAssigned):
			Conflict(c, err.Error())
		case errors.Is(err, service.ErrCompanyNotFound):
			NotFound(c, err.Error())
		default:
			InternalServerError(c, "failed to assign company")
		}
		return
	}

	OK(c, "company assigned successfully", mapper.ToUserResponse(user))
}
