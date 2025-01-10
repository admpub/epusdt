package mdb

import (
	"github.com/dromara/carbon/v2"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint64                `gorm:"column:id;primary_key" json:"id"`
	CreatedAt carbon.TimestampMilli `gorm:"column:created_at" json:"created_at"`
	UpdatedAt carbon.TimestampMilli `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt        `gorm:"index" json:"deleted_at"`
}
