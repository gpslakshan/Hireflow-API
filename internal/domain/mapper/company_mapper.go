package mapper

import (
	"github.com/gpslakshan/hireflow/internal/domain/dto"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
)

func ToCompanyResponse(c entity.Company) dto.CompanyResponse {
	return dto.CompanyResponse{
		ID:          c.ID.String(),
		Name:        c.Name,
		Description: c.Description,
		Industry:    c.Industry,
		Website:     c.Website,
		Location:    c.Location,
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToCompanyResponseList(companies []entity.Company) []dto.CompanyResponse {
	result := make([]dto.CompanyResponse, len(companies))
	for i, c := range companies {
		result[i] = ToCompanyResponse(c)
	}
	return result
}
