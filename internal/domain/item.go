package domain

import (
	"time"

	"github.com/google/uuid"
)

type Item struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string     `gorm:"type:varchar(150);not null" json:"name"`
	Description string     `gorm:"type:text;not null;default:''" json:"description"`
	Stock       int        `gorm:"type:integer;not null;default:0;check:stock >= 0" json:"stock"`
	Price       int64      `gorm:"type:bigint;not null;check:price >= 0" json:"price"`
	CreatedBy   *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"type:timestamptz;not null;default:now()" json:"updated_at"`
}

