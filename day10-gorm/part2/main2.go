package main

import (
	"time"

	"gorm.io/gorm"
)

// Wallet 钱包模型
type Wallet struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID  uint    `gorm:"index" json:"user_id"`
	Address string  `gorm:"uniqueIndex;not null" json:"address"`
	Balance float64 `gorm:"default:0" json:"balance"`
	Network string  `gorm:"not null" json:"network"` // Ethereum, TON, etc.

	// 关联
	User         User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Transactions []Transaction `gorm:"foreignKey:WalletID" json:"transactions,omitempty"`
}

// Transaction 交易模型
type Transaction struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	WalletID uint    `gorm:"index" json:"wallet_id"`
	Hash     string  `gorm:"uniqueIndex;not null" json:"hash"`
	From     string  `gorm:"not null" json:"from"`
	To       string  `gorm:"not null" json:"to"`
	Amount   float64 `gorm:"not null" json:"amount"`
	GasUsed  int64   `json:"gas_used"`
	GasPrice float64 `json:"gas_price"`
	Status   string  `gorm:"default:'pending'" json:"status"` // pending, confirmed, failed
	BlockNum int64   `json:"block_num"`

	// 关联
	Wallet Wallet `gorm:"foreignKey:WalletID" json:"wallet,omitempty"`
}

type NFT struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	TokenID     int    `gorm:"not null" json:"token_id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `gorm:"not null" json:"name"`
	ImageURL    string `json:"image_url"`
	Owner       string `gorm:"index;not null" json:"owner"`
	Contract    string `gorm:"index;not null" json:"contract"`
	Network     string `gorm:"not null" json:"network"`
}
