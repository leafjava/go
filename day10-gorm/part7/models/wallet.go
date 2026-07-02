package models

import "time"

// 用户
type User struct {
	ID       uint     `gorm:"primaryKey" json:"id"`
	Username string   `gorm:"uniqueIndex;not null" json:"username"`
	Wallets  []Wallet `gorm:"foreignKey:UserID" json:"wallets,omitempty"`
}

// 钱包
type Wallet struct {
	ID           uint          `gorm:"primaryKey" json:"id"`
	UserID       uint          `gorm:"index;not null" json:"user_id"`
	Address      string        `gorm:"uniqueIndex;not null" json:"address"`
	Network      string        `gorm:"not null" json:"network"`
	Balance      float64       `gorm:"default:0" json:"balance"`
	Transactions []Transaction `gorm:"foreignKey:WalletID" json:"transactions,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

// 交易记录
type Transaction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	WalletID  uint      `gorm:"index;not null" json:"wallet_id"`
	FromAddr  string    `gorm:"not null" json:"from_addr"`
	ToAddr    string    `gorm:"not null" json:"to_addr"`
	Amount    float64   `gorm:"not null" json:"amount"`
	TxHash    string    `json:"tx_hash"`
	CreatedAt time.Time `json:"created_at"`
}
