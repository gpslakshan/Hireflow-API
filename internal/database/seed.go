package database

import (
	"log"

	"github.com/google/uuid"
	"github.com/gpslakshan/hireflow/internal/config"
	"github.com/gpslakshan/hireflow/internal/domain/entity"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed creates a default admin user if one doesn't already exist.
// Safe to run on every startup — checks before inserting.
func Seed(db *gorm.DB, cfg *config.Config) {
	// Guard — refuse to seed if no password is configured
	if cfg.AdminPassword == "" {
		log.Fatal("ADMIN_PASSWORD is not set in environment — refusing to seed")
	}

	var count int64
	db.Model(&entity.User{}).Where("role = ?", entity.RoleAdmin).Count(&count)

	if count > 0 {
		log.Println("admin user already exists, skipping seed")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), 12)
	if err != nil {
		log.Fatalf("failed to hash seed password: %v", err)
	}

	admin := entity.User{
		ID:           uuid.New(),
		FullName:     "Super Admin",
		Email:        cfg.AdminEmail,
		PasswordHash: string(hash),
		Role:         entity.RoleAdmin,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Fatalf("failed to seed admin user: %v", err)
	}

	if cfg.AppEnv != "production" {
		log.Printf("default admin user created — email: %s", cfg.AdminEmail)
	}
}
