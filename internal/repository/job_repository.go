package repository

import (
	"github.com/google/uuid"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
	"gorm.io/gorm"
)

type JobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{db: db}
}

func (r *JobRepository) Create(job *entity.Job) error {
	return r.db.Create(job).Error
}

// FindAll returns only open jobs for the public feed.
// Preloads Company so the response can include company details.
func (r *JobRepository) FindAll() ([]entity.Job, error) {
	var jobs []entity.Job
	err := r.db.Preload("Company").
		Where("status = ?", entity.JobStatusOpen).
		Find(&jobs).Error
	return jobs, err
}

func (r *JobRepository) FindByID(id uuid.UUID) (*entity.Job, error) {
	var job entity.Job
	err := r.db.Preload("Company").
		Where("id = ?", id).
		First(&job).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *JobRepository) FindByCompanyID(companyID uuid.UUID) ([]entity.Job, error) {
	var jobs []entity.Job
	err := r.db.Where("company_id = ?", companyID).Find(&jobs).Error
	return jobs, err
}

func (r *JobRepository) Update(job *entity.Job) error {
	return r.db.Save(job).Error
}

func (r *JobRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.Job{}, "id = ?", id).Error
}
