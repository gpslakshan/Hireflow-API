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
