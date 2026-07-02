package main

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Email    string `gorm:"uniqueIndex" json:"email"`
	Password string `gorm:"not null" json:"-"`
	Age      int    `json:"age"`
	IsActive bool   `gorm:"default:true" json:"is_active"`
}

func (User) TableName() string {
	return "users"
}
