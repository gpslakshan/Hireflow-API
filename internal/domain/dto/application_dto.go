package dto

type CVUploadURLRequest struct {
	FileName string `json:"file_name" validate:"required"`
}

type CVUploadURLResponse struct {
	UploadURL string `json:"upload_url"`
	CVKey     string `json:"cv_key"`
}

type ApplyJobRequest struct {
	CoverLetter string `json:"cover_letter" validate:"omitempty,max=2000"`
	CVKey       string `json:"cv_key"       validate:"omitempty"`
}

// Why can't a recruiter set status to applied?
// Because applied is the initial state set by the system when a candidate submits.
// Only the system sets it — a recruiter can only move it forward or to rejected. The oneof validator enforces this at the API boundary.
type UpdateApplicationStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=screening interview offer hired rejected"`
}

type ApplicationResponse struct {
	ID            string       `json:"id"`
	Job           JobResponse  `json:"job"`
	Candidate     UserResponse `json:"candidate"`
	CoverLetter   string       `json:"cover_letter"`
	CVDownloadURL string       `json:"cv_download_url,omitempty"`
	Status        string       `json:"status"`
	CreatedAt     string       `json:"created_at"`
	UpdatedAt     string       `json:"updated_at"`
}
