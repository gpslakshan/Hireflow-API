package repository

import (
	"github.com/google/uuid"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
	"gorm.io/gorm"
)

type ApplicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) *ApplicationRepository {
	return &ApplicationRepository{db: db}
}

func (r *ApplicationRepository) Create(app *entity.Application) error {
	return r.db.Create(app).Error
}

// FindByID preloads Job and Candidate for rich responses
func (r *ApplicationRepository) FindByID(id uuid.UUID) (*entity.Application, error) {
	var app entity.Application
	err := r.db.Preload("Job").Preload("Candidate").
		Where("id = ?", id).First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

// FindByCandidate returns all applications submitted by a specific candidate
func (r *ApplicationRepository) FindByCandidate(candidateID uuid.UUID) ([]entity.Application, error) {
	var apps []entity.Application
	err := r.db.Preload("Job").
		Where("candidate_id = ?", candidateID).
		Find(&apps).Error
	return apps, err
}

// FindByJob returns all applications for a specific job
func (r *ApplicationRepository) FindByJob(jobID uuid.UUID) ([]entity.Application, error) {
	var apps []entity.Application
	err := r.db.Preload("Candidate").
		Where("job_id = ?", jobID).
		Find(&apps).Error
	return apps, err
}

// FindByJobAndCandidate is used to check for duplicate applications
func (r *ApplicationRepository) FindByJobAndCandidate(jobID, candidateID uuid.UUID) (*entity.Application, error) {
	var app entity.Application
	err := r.db.Where("job_id = ? AND candidate_id = ?", jobID, candidateID).
		First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *ApplicationRepository) Update(app *entity.Application) error {
	return r.db.Save(app).Error
}

func (r *ApplicationRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.Application{}, "id = ?", id).Error
}
