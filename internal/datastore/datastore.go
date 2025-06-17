package datastore

import (
	"context"
	"time"

	"github.com/DuongQuyen1309/suibottele/internal/db"
	"github.com/DuongQuyen1309/suibottele/internal/model"
)

func CreateBalanceChangeEventTable(ctx context.Context) error {
	_, err := db.DB.NewCreateTable().
		Model((*model.BalanceChangeEvent)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}
	_, err = db.DB.NewCreateIndex().Model((*model.BalanceChangeEvent)(nil)).
		Index("idx_sui_transaction_hash_address_token").
		Unique().
		Column("transaction_hash", "wallet_address", "token").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}
	_, err = db.DB.NewCreateIndex().Model((*model.BalanceChangeEvent)(nil)).
		Index("idx_sui_transaction_hash").
		Column("transaction_hash").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}
	_, err = db.DB.NewCreateIndex().Model((*model.BalanceChangeEvent)(nil)).
		Index("idx_sui_token_amount").
		Column("token", "amount").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}
	_, err = db.DB.NewCreateIndex().Model((*model.BalanceChangeEvent)(nil)).
		Index("idx_sui_created").
		Column("created_at").
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}
func InsertDB(wallet string, amount float64, rawAmount string, digest string, symbol string, timestamp time.Time, ctx context.Context) error {
	_, err := db.DB.NewInsert().Model(&model.BalanceChangeEvent{
		WalletAddress:   wallet,
		Amount:          amount,
		RawAmount:       rawAmount,
		Token:           symbol,
		TransactionHash: digest,
		CreatedAt:       timestamp,
	}).On("CONFLICT (transaction_hash, wallet_address, token) DO NOTHING").Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func CalculaterReceivedAmount(coinType string, ctx context.Context) (float64, error) {
	var totalAmount float64
	err := db.DB.NewSelect().
		ColumnExpr("SUM(amount)").
		Model((*model.BalanceChangeEvent)(nil)).
		Where("token = ?", coinType).
		Where("amount > 0").
		Scan(ctx, &totalAmount)
	if err != nil {
		return 0, err
	}
	return totalAmount, nil
}
func CalculaterSentAmount(coinType string, ctx context.Context) (float64, error) {
	var totalAmount float64
	err := db.DB.NewSelect().
		ColumnExpr("SUM(amount)").
		Model((*model.BalanceChangeEvent)(nil)).
		Where("token = ?", coinType).
		Where("amount < 0").
		Scan(ctx, &totalAmount)
	if err != nil {
		return 0, err
	}
	return totalAmount, nil
}

func GetBalanceChangeEvents(hash string, offset int, limit int, ctx context.Context) (*[]model.BalanceChangeEvent, error) {
	var transaction []model.BalanceChangeEvent
	err := db.DB.NewSelect().
		Model(&transaction).
		Where("transaction_hash = ?", hash).
		Offset(offset).
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func GetBalanceChangeEventInRange(fromDate time.Time, toDate time.Time, offset int, limit int, ctx context.Context) (*[]model.BalanceChangeEvent, error) {
	var transaction []model.BalanceChangeEvent
	err := db.DB.NewSelect().
		Model(&transaction).
		Where("created_at >= ?", fromDate).
		Where("created_at <= ?", toDate).
		Offset(offset).
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}
