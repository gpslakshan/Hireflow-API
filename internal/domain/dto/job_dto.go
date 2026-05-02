package dto

type CreateJobRequest struct {
	Title       string `json:"title"       validate:"required,min=2,max=150" example:"Senior Go Engineer"`
	Description string `json:"description" validate:"required"                example:"We are looking for an experienced Go engineer."`
	Location    string `json:"location"    validate:"required,max=150"        example:"Colombo, Sri Lanka"`
	JobType     string `json:"job_type"    validate:"required,oneof=full_time part_time contract internship" example:"full_time"`
}

type UpdateJobRequest struct {
	Title       string `json:"title"       validate:"omitempty,min=2,max=150" example:"Senior Go Engineer (Updated)"`
	Description string `json:"description" validate:"omitempty"                example:"Updated description."`
	Location    string `json:"location"    validate:"omitempty,max=150"        example:"Remote"`
	JobType     string `json:"job_type"    validate:"omitempty,oneof=full_time part_time contract internship" example:"contract"`
}

type JobResponse struct {
	ID          string          `json:"id"          example:"b2c3d4e5-..."`
	Title       string          `json:"title"       example:"Senior Go Engineer"`
	Description string          `json:"description" example:"We are looking for an experienced Go engineer."`
	Location    string          `json:"location"    example:"Colombo, Sri Lanka"`
	JobType     string          `json:"job_type"    example:"full_time"`
	Status      string          `json:"status"      example:"open"`
	Company     CompanyResponse `json:"company"`
	CreatedAt   string          `json:"created_at"  example:"2026-05-01T12:00:00Z"`
}
