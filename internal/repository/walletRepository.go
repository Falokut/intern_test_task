package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Falokut/intern_test_task/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type walletRepository struct {
	db             *sqlx.DB
	logger         *logrus.Logger
	defaultBalance float32
}

func NewWalletRepository(logger *logrus.Logger, db *sqlx.DB, defaultBalance float32) *walletRepository {
	return &walletRepository{
		logger:         logger,
		db:             db,
		defaultBalance: defaultBalance,
	}
}

const (
	walletsTableName     = "wallets"
	historyDataTableName = "history"
)

func (r *walletRepository) CreateWallet(ctx context.Context) (models.Wallet, error) {
	query := fmt.Sprintf(`INSERT INTO %s (balance) VALUES($1) RETURNING id, balance;`, walletsTableName)
	var wallet models.Wallet
	err := r.db.GetContext(ctx, &wallet, query, r.defaultBalance)
	if err != nil {
		r.logger.Errorf("error: %s query: %s", err.Error(), query)
		return models.Wallet{}, err
	}

	return wallet, nil
}

func (r *walletRepository) GetWalletBalance(ctx context.Context, id string) (float32, error) {
	query := fmt.Sprintf("SELECT balance FROM %s WHERE id=$1", walletsTableName)
	var balance float32
	err := r.db.GetContext(ctx, &balance, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0.0, ErrWalletNotFound
	}
	if err != nil {
		r.logger.Errorf("error: %s query: %s", err.Error(), query)
		return 0.0, err
	}

	return balance, nil
}

func (r *walletRepository) FundTranswer(ctx context.Context, fromID, toID string, amount float32) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := fmt.Sprintf("UPDATE %s SET balance=balance-$1 WHERE id=$2", walletsTableName)
	_, err = tx.ExecContext(ctx, query, amount, fromID)
	if err != nil {
		r.logger.Errorf("error: %s query: %s", err.Error(), query)
		return err
	}

	query = fmt.Sprintf("UPDATE %s SET balance=balance+$1 WHERE id=$2", walletsTableName)
	_, err = tx.ExecContext(ctx, query, amount, toID)
	if err != nil {
		r.logger.Errorf("error: %s query: %s", err.Error(), query)
		return err
	}

	query = fmt.Sprintf("INSERT INTO %s (from_wallet,to_wallet,amount) VALUES($1,$2,$3)", historyDataTableName)
	_, err = tx.ExecContext(ctx, query, fromID, toID, amount)
	if err != nil {
		r.logger.Errorf("error: %s query: %s", err.Error(), query)
		return err
	}

	tx.Commit()
	return nil
}

func (r *walletRepository) GetWalletHistory(ctx context.Context, id string) ([]models.Transaction, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE from_wallet=$1 OR to_wallet=$1 ORDER BY time", historyDataTableName)
	var history []models.Transaction
	err := r.db.SelectContext(ctx, &history, query, id)
	if err != nil {
		r.logger.Errorf("error: %s query: %s", err.Error(), query)
		return []models.Transaction{}, err
	}

	return history, nil
}

func (r *walletRepository) IsWalletExists(ctx context.Context, id string) (bool, error) {
	query := fmt.Sprintf("SELECT id FROM %s WHERE id=$1", walletsTableName)

	err := r.db.GetContext(ctx, &id, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}
