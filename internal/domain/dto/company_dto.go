package dto

type CreateCompanyRequest struct {
	Name        string `json:"name"        validate:"required,min=2,max=150"`
	Description string `json:"description" validate:"required"`
	Industry    string `json:"industry"    validate:"required,max=100"`
	Website     string `json:"website"     validate:"omitempty,url"`
	Location    string `json:"location"    validate:"required,max=150"`
}

type UpdateCompanyRequest struct {
	Name        string `json:"name"        validate:"omitempty,min=2,max=150"`
	Description string `json:"description" validate:"omitempty"`
	Industry    string `json:"industry"    validate:"omitempty,max=100"`
	Website     string `json:"website"     validate:"omitempty,url"`
	Location    string `json:"location"    validate:"omitempty,max=150"`
}

type CompanyResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Industry    string `json:"industry"`
	Website     string `json:"website"`
	Location    string `json:"location"`
	CreatedAt   string `json:"created_at"`
}
