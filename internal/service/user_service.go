package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/gpslakshan/hireflow/internal/domain/dto"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
	"github.com/gpslakshan/hireflow/internal/repository"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrNotARecruiter   = errors.New("user is not a recruiter")
	ErrAlreadyAssigned = errors.New("recruiter is already assigned to a company")
)

type UserService struct {
	userRepo    repository.UserRepo
	companyRepo repository.CompanyRepo
}

func NewUserService(
	userRepo repository.UserRepo,
	companyRepo repository.CompanyRepo,
) *UserService {
	return &UserService{
		userRepo:    userRepo,
		companyRepo: companyRepo,
	}
}

// AssignCompany links a recruiter to a company.
// Only admins should call this — enforced at the handler/middleware level.
func (s *UserService) AssignCompany(userID string, req dto.AssignCompanyRequest) (entity.User, error) {
	// 1. Fetch the user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return entity.User{}, ErrUserNotFound
	}

	// 2. Business rule: only recruiters can be assigned to a company
	if user.Role != entity.RoleRecruiter {
		return entity.User{}, ErrNotARecruiter
	}

	// 3. Business rule: prevent re-assigning without explicit unassign
	if user.CompanyID != nil {
		return entity.User{}, ErrAlreadyAssigned
	}

	// 4. Verify the company exists
	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		return entity.User{}, fmt.Errorf("invalid company id: %w", err)
	}

	_, err = s.companyRepo.FindByID(companyID)
	if err != nil {
		return entity.User{}, ErrCompanyNotFound
	}

	// 5. Assign and save
	user.CompanyID = &companyID
	if err := s.userRepo.Update(user); err != nil {
		return entity.User{}, fmt.Errorf("failed to assign company: %w", err)
	}

	return *user, nil
}
