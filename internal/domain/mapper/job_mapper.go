package mapper

import (
	"github.com/gpslakshan/hireflow/internal/domain/dto"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
)

func ToJobResponse(j entity.Job) dto.JobResponse {
	return dto.JobResponse{
		ID:          j.ID.String(),
		Title:       j.Title,
		Description: j.Description,
		Location:    j.Location,
		JobType:     string(j.JobType),
		Status:      string(j.Status),
		Company:     ToCompanyResponse(j.Company),
		CreatedAt:   j.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToJobResponseList(jobs []entity.Job) []dto.JobResponse {
	result := make([]dto.JobResponse, len(jobs))
	for i, j := range jobs {
		result[i] = ToJobResponse(j)
	}
	return result
}
