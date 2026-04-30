package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string
type JobType string
type JobStatus string
type ApplicationStatus string

const (
	RoleAdmin     Role = "admin"
	RoleRecruiter Role = "recruiter"
	RoleCandidate Role = "candidate"
)

// User represents the users table.
type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	FullName     string     `gorm:"type:varchar(100);not null"`
	Email        string     `gorm:"type:varchar(150);uniqueIndex;not null"`
	PasswordHash string     `gorm:"type:varchar(255);not null"`
	Role         Role       `gorm:"type:varchar(20);not null"`
	CompanyID    *uuid.UUID `gorm:"type:uuid"`            // nullable — only set for recruiters
	Company      *Company   `gorm:"foreignKey:CompanyID"` // GORM association
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"` // enables soft delete automatically
}

// The BeforeCreate Hook - This is a GORM lifecycle hook.
// BeforeCreate hook — GORM calls this automatically before every INSERT.
// In GORM, tx *gorm.DB represents a pointer to a GORM Database instance.
// tx is the active transaction that GORM is using to perform the insert operation.
// Generates a UUID so we never rely on the database to generate it.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
