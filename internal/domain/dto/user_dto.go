package dto

type AssignCompanyRequest struct {
	CompanyID string `json:"company_id" validate:"required,uuid4" example:"a1b2c3d4-..."`
}
