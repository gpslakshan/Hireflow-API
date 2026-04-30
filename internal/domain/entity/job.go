package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	JobTypeFull     JobType = "full_time"
	JobTypePart     JobType = "part_time"
	JobTypeContract JobType = "contract"
	JobTypeIntern   JobType = "internship"

	JobStatusOpen   JobStatus = "open"
	JobStatusClosed JobStatus = "closed"
)

type Job struct {
	ID           uuid.UUID     `gorm:"type:uuid;primaryKey"`
	CompanyID    uuid.UUID     `gorm:"type:uuid;not null;index"`
	Company      Company       `gorm:"foreignKey:CompanyID"`
	PostedBy     uuid.UUID     `gorm:"type:uuid;not null"`
	Poster       User          `gorm:"foreignKey:PostedBy"`
	Title        string        `gorm:"type:varchar(150);not null"`
	Description  string        `gorm:"type:text;not null"`
	Location     string        `gorm:"type:varchar(150)"`
	JobType      JobType       `gorm:"type:varchar(20);not null"`
	Status       JobStatus     `gorm:"type:varchar(20);not null;default:'open'"`
	Applications []Application `gorm:"foreignKey:JobID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (j *Job) BeforeCreate(tx *gorm.DB) error {
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	return nil
}
