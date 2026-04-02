package migrations

import (
	"gorm.io/gorm"
)

func EnablePostGIS(db *gorm.DB) error {
	return db.Exec("CREATE EXTENSION IF NOT EXISTS postgis").Error
}

func CreateSpatialIndexes(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_companies_location 
		ON companies 
		USING GIST (ST_MakePoint(lng, lat))
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_dishes_company_id 
		ON dishes (company_id)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_swipe_actions_session_dish 
		ON swipe_actions (session_id, dish_id)
	`).Error; err != nil {
		return err
	}

	return nil
}

func RunMigrations(db *gorm.DB) error {
	if err := EnablePostGIS(db); err != nil {
		return err
	}

	return CreateSpatialIndexes(db)
}
