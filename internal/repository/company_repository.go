package repository

import (
	"github.com/google/uuid"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
	"gorm.io/gorm"
)

type CompanyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

func (r *CompanyRepository) Create(company *entity.Company) error {
	return r.db.Create(company).Error
}

func (r *CompanyRepository) FindAll() ([]entity.Company, error) {
	var companies []entity.Company
	err := r.db.Find(&companies).Error
	return companies, err
}

// FindByID returns a pointer to the company entity to support two main goals:
//  1. Semantic Clarity: It allows returning 'nil' to explicitly signal the record wasn't found.
//  2. Mutability: It provides the caller with a direct reference to the object in memory.
//     This allows the Service layer to modify fields directly on the retrieved instance
//     and pass that same instance back to the Update method without unnecessary memory copying.
func (r *CompanyRepository) FindByID(id uuid.UUID) (*entity.Company, error) {
	var company entity.Company
	err := r.db.Where("id = ?", id).First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *CompanyRepository) Update(company *entity.Company) error {
	return r.db.Save(company).Error
}

func (r *CompanyRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&entity.Company{}, "id = ?", id).Error
}
