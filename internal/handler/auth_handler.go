package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gpslakshan/hireflow/internal/domain/dto"
	"github.com/gpslakshan/hireflow/internal/domain/mapper"
	"github.com/gpslakshan/hireflow/internal/service"
)

type AuthHandler struct {
	authService service.AuthServiceInterface
	validate    *validator.Validate
}

func NewAuthHandler(authService service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validate:    validator.New(),
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new candidate or recruiter account
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.RegisterRequest  true  "Registration details"
// @Success      201      {object}  handler.APIResponse{data=dto.UserResponse}
// @Failure      400      {object}  handler.APIResponse
// @Failure      409      {object}  handler.APIResponse
// @Failure      422      {object}  handler.APIResponse{errors=map[string]string}
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, formatValidationErrors(err))
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			Conflict(c, err.Error())
			return
		}
		InternalServerError(c, "registration failed")
		return
	}

	Created(c, "user registered successfully", mapper.ToUserResponse(user))
}

// Login godoc
// @Summary      Login
// @Description  Authenticate with email and password, receive a JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.LoginRequest  true  "Login credentials"
// @Success      200      {object}  handler.APIResponse{data=dto.AuthResponse}
// @Failure      400      {object}  handler.APIResponse
// @Failure      401      {object}  handler.APIResponse
// @Failure      422      {object}  handler.APIResponse{errors=map[string]string}
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, formatValidationErrors(err))
		return
	}

	token, user, err := h.authService.Login(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			Unauthorized(c, err.Error())
			return
		}
		InternalServerError(c, "login failed")
		return
	}

	OK(c, "login successful", dto.AuthResponse{
		Token: token,
		User:  mapper.ToUserResponse(user),
	})
}

// Me godoc
// @Summary      Get current user
// @Description  Returns the profile of the currently authenticated user
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  handler.APIResponse{data=dto.UserResponse}
// @Failure      401  {object}  handler.APIResponse
// @Failure      404  {object}  handler.APIResponse
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.authService.GetByID(userID.(string))
	if err != nil {
		NotFound(c, "user not found")
		return
	}

	OK(c, "user retrieved successfully", mapper.ToUserResponse(*user))
}
