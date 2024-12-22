package transaction

import (
	"context"
	"database/sql"
	"fmt"
)

// TxManager gerencia transações do banco de dados
type TxManager struct {
	db *sql.DB
}

// NewTxManager cria uma nova instância do TxManager
func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{db: db}
}

// RunInTransaction executa uma função dentro de uma transação
func (tm *TxManager) RunInTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %v", err)
	}
	
	// Garante que a transação será finalizada
	defer func() {
		if p := recover(); p != nil {
			// Em caso de panic, faz rollback e re-panic
			_ = tx.Rollback()
			panic(p)
		}
	}()
	
	// Executa a função dentro da transação
	if err := fn(tx); err != nil {
		// Em caso de erro, faz rollback
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("erro ao fazer rollback: %v (erro original: %v)", rbErr, err)
		}
		return err
	}
	
	// Commit da transação
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erro ao fazer commit: %v", err)
	}
	
	return nil
} 