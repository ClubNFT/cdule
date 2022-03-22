package model

import (
	"gorm.io/gorm"
)

// DB gorm DB
var DB *gorm.DB

// CduleRepos repositories
var CduleRepos *Repositories

// Repositories struct
type Repositories struct {
	CduleRepository CduleRepository
	DB              *gorm.DB
}

// ConnectDataBase to create a database connection
func ConnectDataBase(db *gorm.DB) {
	Migrate(db)
	DB = db
	// Initialise CduleRepositories
	CduleRepos = &Repositories{
		CduleRepository: NewCduleRepository(db),
		DB:              db,
	}
}

// Migrate database schema
func Migrate(db *gorm.DB) {
	db.AutoMigrate(&Job{})
	db.AutoMigrate(&JobHistory{})
	db.AutoMigrate(&Schedule{})
	db.AutoMigrate(&Worker{})
}
