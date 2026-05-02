package dto

type CVUploadURLRequest struct {
	FileName string `json:"file_name" validate:"required" example:"my-cv.pdf"`
}

type CVUploadURLResponse struct {
	UploadURL string `json:"upload_url" example:"https://hireflow-cvs.s3.amazonaws.com/cvs/...?X-Amz-Signature=..."`
	CVKey     string `json:"cv_key"     example:"cvs/da9d5716-4979-4f93-980a-9dfcf00cd02a/my-cv.pdf"`
}

type ApplyJobRequest struct {
	CoverLetter string `json:"cover_letter" validate:"omitempty,max=2000" example:"I am very excited to apply."`
	CVKey       string `json:"cv_key"       validate:"omitempty"          example:"cvs/da9d5716-4979-4f93-980a-9dfcf00cd02a/my-cv.pdf"`
}

// Why can't a recruiter set status to applied?
// Because applied is the initial state set by the system when a candidate submits.
// Only the system sets it — a recruiter can only move it forward or to rejected. The oneof validator enforces this at the API boundary.
type UpdateApplicationStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=screening interview offer hired rejected" example:"screening"`
}

type ApplicationResponse struct {
	ID            string       `json:"id"              example:"c3d4e5f6-..."`
	Job           JobResponse  `json:"job"`
	Candidate     UserResponse `json:"candidate"`
	CoverLetter   string       `json:"cover_letter"    example:"I am very excited to apply."`
	CVDownloadURL string       `json:"cv_download_url,omitempty" example:"https://s3.amazonaws.com/..."`
	Status        string       `json:"status"          example:"applied"`
	CreatedAt     string       `json:"created_at"      example:"2026-05-01T12:00:00Z"`
	UpdatedAt     string       `json:"updated_at"      example:"2026-05-01T13:00:00Z"`
}
