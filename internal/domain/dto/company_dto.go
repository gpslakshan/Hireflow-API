package dto

type CreateCompanyRequest struct {
	Name        string `json:"name"        validate:"required,min=2,max=150" example:"TechCorp Lanka"`
	Description string `json:"description" validate:"required"                example:"A leading software company."`
	Industry    string `json:"industry"    validate:"required,max=100"        example:"Technology"`
	Website     string `json:"website"     validate:"omitempty,url"           example:"https://techcorplanka.com"`
	Location    string `json:"location"    validate:"required,max=150"        example:"Colombo, Sri Lanka"`
}

type UpdateCompanyRequest struct {
	Name        string `json:"name"        validate:"omitempty,min=2,max=150" example:"TechCorp Lanka"`
	Description string `json:"description" validate:"omitempty"                example:"Updated description."`
	Industry    string `json:"industry"    validate:"omitempty,max=100"        example:"Technology"`
	Website     string `json:"website"     validate:"omitempty,url"            example:"https://techcorplanka.com"`
	Location    string `json:"location"    validate:"omitempty,max=150"        example:"Colombo 03, Sri Lanka"`
}

type CompanyResponse struct {
	ID          string `json:"id"          example:"a1b2c3d4-..."`
	Name        string `json:"name"        example:"TechCorp Lanka"`
	Description string `json:"description" example:"A leading software company."`
	Industry    string `json:"industry"    example:"Technology"`
	Website     string `json:"website"     example:"https://techcorplanka.com"`
	Location    string `json:"location"    example:"Colombo, Sri Lanka"`
	CreatedAt   string `json:"created_at"  example:"2026-05-01T12:00:00Z"`
}
