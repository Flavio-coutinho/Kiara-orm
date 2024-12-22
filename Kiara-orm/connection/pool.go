package connection

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

// Pool representa o pool de conexões com o banco de dados
type Pool struct {
	db     *sql.DB
	config *Config
	mu     sync.RWMutex
}

// NewPool cria uma nova instância do pool de conexões
func NewPool(config *Config) (*Pool, error) {
	pool := &Pool{
		config: config,
	}
	
	if err := pool.connect(); err != nil {
		return nil, fmt.Errorf("falha ao criar pool de conexões: %v", err)
	}
	
	return pool, nil
}

// connect estabelece a conexão com o banco de dados
func (p *Pool) connect() error {
	db, err := sql.Open(string(p.config.Type), p.config.DSN())
	if err != nil {
		return err
	}
	
	// Configura o pool de conexões
	db.SetMaxOpenConns(p.config.MaxOpenConns)
	db.SetMaxIdleConns(p.config.MaxIdleConns)
	db.SetConnMaxLifetime(p.config.ConnMaxLifetime)
	
	// Testa a conexão
	if err := db.Ping(); err != nil {
		return fmt.Errorf("falha ao conectar ao banco de dados: %v", err)
	}
	
	p.db = db
	return nil
}

// GetDB retorna a conexão com o banco de dados
func (p *Pool) GetDB() *sql.DB {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.db
}

// Begin inicia uma nova transação
func (p *Pool) Begin(ctx context.Context) (*sql.Tx, error) {
	return p.db.BeginTx(ctx, nil)
}

// Close fecha todas as conexões do pool
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// Ping verifica se a conexão com o banco está ativa
func (p *Pool) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Stats retorna estatísticas sobre o pool de conexões
func (p *Pool) Stats() sql.DBStats {
	return p.db.Stats()
}
