package model

import (
	"time"

	"github.com/uptrace/bun"
)

type SuiTransaction struct {
	bun.BaseModel `bun:"table:suichains"`

	Id              int       `bun:"id,pk,autoincrement" json:"id"`
	WalletAddress   string    `bun:"wallet_address" json:"walletAddress"`
	Amount          float64   `bun:"amount" json:"amount"`
	RawAmount       string    `bun:"raw_amount" json:"rawAmount"`
	Token           string    `bun:"token" json:"token"`
	TransactionHash string    `bun:"transaction_hash" json:"transactionHash"`
	CreatedAt       time.Time `bun:"created_at" json:"createdAt"`
}
