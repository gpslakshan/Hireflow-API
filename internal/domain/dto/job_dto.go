package dto

type CreateJobRequest struct {
	Title       string `json:"title"       validate:"required,min=2,max=150"`
	Description string `json:"description" validate:"required"`
	Location    string `json:"location"    validate:"required,max=150"`
	JobType     string `json:"job_type"    validate:"required,oneof=full_time part_time contract internship"`
}

type UpdateJobRequest struct {
	Title       string `json:"title"       validate:"omitempty,min=2,max=150"`
	Description string `json:"description" validate:"omitempty"`
	Location    string `json:"location"    validate:"omitempty,max=150"`
	JobType     string `json:"job_type"    validate:"omitempty,oneof=full_time part_time contract internship"`
}

type JobResponse struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Location    string          `json:"location"`
	JobType     string          `json:"job_type"`
	Status      string          `json:"status"`
	Company     CompanyResponse `json:"company"`
	CreatedAt   string          `json:"created_at"`
}
