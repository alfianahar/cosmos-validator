package models

import "time"

// MigrationHistory tracks database migration events
type MigrationHistory struct {
	ID           uint      `gorm:"primaryKey"`
	MigratedAt   time.Time `gorm:"autoCreateTime;index"`
	Models       string    `gorm:"type:text"`
	Status       string    `gorm:"type:varchar(50);index"`
	ErrorMessage string    `gorm:"type:text"`
}
