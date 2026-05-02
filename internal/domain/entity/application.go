package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApplicationStatus string

const (
	StatusApplied   ApplicationStatus = "applied"
	StatusScreening ApplicationStatus = "screening"
	StatusInterview ApplicationStatus = "interview"
	StatusOffer     ApplicationStatus = "offer"
	StatusHired     ApplicationStatus = "hired"
	StatusRejected  ApplicationStatus = "rejected"
)

type Application struct {
	ID          uuid.UUID         `gorm:"type:uuid;primaryKey"`
	JobID       uuid.UUID         `gorm:"type:uuid;not null;index"`
	Job         Job               `gorm:"foreignKey:JobID"`
	CandidateID uuid.UUID         `gorm:"type:uuid;not null;index"`
	Candidate   User              `gorm:"foreignKey:CandidateID"`
	CoverLetter string            `gorm:"type:text"`
	CVKey       string            `gorm:"type:varchar(500)"` // ← S3 object key
	Status      ApplicationStatus `gorm:"type:varchar(20);not null;default:'applied'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// No DeletedAt — applications are hard deleted (withdrawn)
}

func (a *Application) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
