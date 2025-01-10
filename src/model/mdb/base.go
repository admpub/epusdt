package mdb

import (
	"database/sql"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint64         `gorm:"column:id;primary_key" json:"id"`
	CreatedAt sql.NullTime   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt sql.NullTime   `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
