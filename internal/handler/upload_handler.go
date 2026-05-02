package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gpslakshan/hireflow/internal/domain/dto"
	"github.com/gpslakshan/hireflow/internal/service"
)

type UploadHandler struct {
	uploadService service.UploadServiceInterface
	validate      *validator.Validate
}

func NewUploadHandler(uploadService service.UploadServiceInterface) *UploadHandler {
	return &UploadHandler{
		uploadService: uploadService,
		validate:      validator.New(),
	}
}

// GenerateCVUploadURL godoc
// @Summary      Get a CV upload URL
// @Description  Returns a pre-signed S3 URL for uploading a PDF CV directly to S3.
//
//	After uploading, include the returned cv_key when submitting an application.
//
// @Tags         Uploads
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.CVUploadURLRequest  true  "File name"
// @Success      200      {object}  handler.APIResponse{data=dto.CVUploadURLResponse}
// @Failure      400      {object}  handler.APIResponse
// @Failure      401      {object}  handler.APIResponse
// @Failure      403      {object}  handler.APIResponse
// @Router       /uploads/cv-upload-url [post]
func (h *UploadHandler) GenerateCVUploadURL(c *gin.Context) {
	candidateID, err := extractUserID(c)
	if err != nil {
		Unauthorized(c, "unauthorized")
		return
	}

	var req dto.CVUploadURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		ValidationError(c, formatValidationErrors(err))
		return
	}

	result, err := h.uploadService.GenerateCVUploadURL(
		c.Request.Context(),
		candidateID.String(),
		req.FileName,
	)
	if err != nil {
		if err.Error() == service.ErrInvalidFileType.Error() {
			BadRequest(c, err.Error())
			return
		}
		InternalServerError(c, "failed to generate upload URL")
		return
	}

	OK(c, "upload URL generated successfully", result)
}
