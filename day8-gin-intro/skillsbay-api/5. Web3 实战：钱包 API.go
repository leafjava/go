package main

type Wallet struct {
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
	Network string  `json:"network"`
}

type TransferRequest struct {
	To     string  `json:"to" binding:"required"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
}
