package mapper

import (
	"github.com/gpslakshan/hireflow/internal/domain/dto"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
)

// ToUserResponse converts a User entity to a safe response DTO.
// This is the only place that decides what user fields are exposed publicly.
func ToUserResponse(u entity.User) dto.UserResponse {
	// In our database entity, CompanyID is a pointer (*uuid.UUID), meaning it can be nil.
	// The Problem: If we try to call .String() on a nil pointer, the program will crash (panic).
	// The Solution: This code safely checks if the pointer exists. If it’s nil, it defaults to an empty string "". If it exists, it converts the UUID to a readable string.
	companyID := ""
	if u.CompanyID != nil {
		companyID = u.CompanyID.String()
	}

	return dto.UserResponse{
		ID:        u.ID.String(),
		FullName:  u.FullName,
		Email:     u.Email,
		Role:      string(u.Role),
		CompanyID: companyID,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
