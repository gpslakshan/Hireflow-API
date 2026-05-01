package mapper

import (
	"github.com/gpslakshan/hireflow/internal/domain/dto"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
)

func ToApplicationResponse(a entity.Application) dto.ApplicationResponse {
	return dto.ApplicationResponse{
		ID:          a.ID.String(),
		Job:         ToJobResponse(a.Job),
		Candidate:   ToUserResponse(a.Candidate),
		CoverLetter: a.CoverLetter,
		Status:      string(a.Status),
		CreatedAt:   a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   a.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToApplicationResponseList(apps []entity.Application) []dto.ApplicationResponse {
	result := make([]dto.ApplicationResponse, len(apps))
	for i, a := range apps {
		result[i] = ToApplicationResponse(a)
	}
	return result
}
