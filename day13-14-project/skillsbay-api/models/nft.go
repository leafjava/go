package models

import "gorm.io/gorm"

type NFT struct {
	gorm.Model
	UserID      uint   `gorm:"index" json:"user_id"`
	TokenID     string `gorm:"uniqueIndex;not null" json:"token_id"`
	ContractAddress string `gorm:"not null" json:"contract_address"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Network     string `gorm:"not null" json:"network"`
	Owner       User   `gorm:"foreignKey:UserID" json:"owner,omitempty"`
}
