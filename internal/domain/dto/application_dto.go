package dto

type ApplyJobRequest struct {
	CoverLetter string `json:"cover_letter" validate:"omitempty,max=2000"`
}

// Why can't a recruiter set status to applied?
// Because applied is the initial state set by the system when a candidate submits.
// Only the system sets it — a recruiter can only move it forward or to rejected. The oneof validator enforces this at the API boundary.
type UpdateApplicationStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=screening interview offer hired rejected"`
}

// ApplicationResponse is the full application shape returned to clients
type ApplicationResponse struct {
	ID          string       `json:"id"`
	Job         JobResponse  `json:"job"`
	Candidate   UserResponse `json:"candidate"`
	CoverLetter string       `json:"cover_letter"`
	Status      string       `json:"status"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
}
