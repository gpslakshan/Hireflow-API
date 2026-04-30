package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Company struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name        string    `gorm:"type:varchar(150);uniqueIndex;not null"`
	Description string    `gorm:"type:text"`
	Industry    string    `gorm:"type:varchar(100)"`
	Website     string    `gorm:"type:varchar(255)"`
	Location    string    `gorm:"type:varchar(150)"`
	Jobs        []Job     `gorm:"foreignKey:CompanyID"` // one company → many jobs
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (c *Company) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
