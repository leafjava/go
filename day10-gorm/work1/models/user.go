package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"uniqueIndex;not null;size:50" json:"username" binding:"required,min=3,max=20"`
	Email     string         `gorm:"uniqueIndex;not null;size:100" json:"email" binding:"required,email"`
	Age       int            `gorm:"not null;default:0" json:"age" binding:"gte=0,lte=150"`
	Role      string         `gorm:"not null;default:user;size:20" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
