package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseEntity struct {
	ID 			uuid.UUID		`gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt 	time.Time		`gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt 	time.Time		`gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt 	gorm.DeletedAt	`gorm:"index" json:"deleted_at,omitempty"`
}

func (base *BaseEntity) BeforeCreate(tx *gorm.DB) error {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return  nil
}