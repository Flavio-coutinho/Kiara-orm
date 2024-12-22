package schema

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	
	"github.com/Flavio-coutinho/Kiara-orm/dialect"
)

// Migration representa uma migração do banco de dados
type Migration struct {
	ID        int64
	Name      string
	Timestamp time.Time
	Applied   bool
}

// Migrator é responsável por gerenciar as migrações
type Migrator struct {
	db      *sql.DB
	dialect dialect.Dialect
	parser  *Parser
}

// NewMigrator cria uma nova instância do Migrator
func NewMigrator(db *sql.DB, dialect dialect.Dialect) *Migrator {
	return &Migrator{
		db:      db,
		dialect: dialect,
		parser:  NewParser(),
	}
}

// ensureMigrationTable garante que a tabela de migrações existe
func (m *Migrator) ensureMigrationTable(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			applied BOOLEAN NOT NULL DEFAULT FALSE
		)
	`, m.dialect.Quote("migrations"))
	
	_, err := m.db.ExecContext(ctx, query)
	return err
}

// AutoMigrate cria ou atualiza tabelas baseado nas structs
func (m *Migrator) AutoMigrate(ctx context.Context, models ...interface{}) error {
	if err := m.ensureMigrationTable(ctx); err != nil {
		return fmt.Errorf("erro ao criar tabela de migrações: %v", err)
	}
	
	for _, model := range models {
		if err := m.migrateModel(ctx, model); err != nil {
			return err
		}
	}
	
	return nil
}

// migrateModel migra uma única struct
func (m *Migrator) migrateModel(ctx context.Context, model interface{}) error {
	mapping, err := m.parser.Parse(model)
	if err != nil {
		return err
	}
	
	// Verifica se a tabela existe
	exists, err := m.tableExists(ctx, mapping.TableName)
	if err != nil {
		return err
	}
	
	if !exists {
		// Cria a tabela se não existir
		return m.createTable(ctx, *mapping)
	}
	
	// Atualiza a tabela existente
	return m.updateTable(ctx, *mapping)
}

// tableExists verifica se uma tabela existe
func (m *Migrator) tableExists(ctx context.Context, tableName string) (bool, error) {
	var query string
	switch m.dialect.(type) {
	case *dialect.PostgreSQL:
		query = `
			SELECT EXISTS (
				SELECT FROM pg_tables
				WHERE schemaname = 'public'
				AND tablename = $1
			)
		`
	case *dialect.MySQL:
		query = `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = DATABASE()
				AND table_name = ?
			)
		`
	case *dialect.SQLite:
		query = `
			SELECT EXISTS (
				SELECT 1 FROM sqlite_master
				WHERE type = 'table'
				AND name = ?
			)
		`
	default:
		return false, fmt.Errorf("dialeto não suportado")
	}
	
	var exists bool
	err := m.db.QueryRowContext(ctx, query, tableName).Scan(&exists)
	return exists, err
}

// createTable cria uma nova tabela
func (m *Migrator) createTable(ctx context.Context, mapping TableMapping) error {
	query := m.dialect.CreateTableSQL(mapping)
	
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela %s: %v", mapping.TableName, err)
	}
	
	// Registra a migração
	return m.recordMigration(ctx, fmt.Sprintf("create_table_%s", mapping.TableName))
}

// updateTable atualiza uma tabela existente
func (m *Migrator) updateTable(ctx context.Context, mapping TableMapping) error {
	// Obtém informações da tabela atual
	currentColumns, err := m.getTableColumns(ctx, mapping.TableName)
	if err != nil {
		return err
	}
	
	// Compara e atualiza colunas
	for _, field := range mapping.Fields {
		if _, exists := currentColumns[field.Name]; !exists {
			// Adiciona coluna nova
			query := m.dialect.AddColumnSQL(mapping.TableName, field)
			if _, err := m.db.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("erro ao adicionar coluna %s: %v", field.Name, err)
			}
			
			// Registra a migração
			if err := m.recordMigration(ctx, fmt.Sprintf("add_column_%s_%s", 
				mapping.TableName, field.Name)); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// getTableColumns obtém as colunas existentes de uma tabela
func (m *Migrator) getTableColumns(ctx context.Context, tableName string) (map[string]bool, error) {
	var query string
	switch m.dialect.(type) {
	case *dialect.PostgreSQL:
		query = `
			SELECT column_name
			FROM information_schema.columns
			WHERE table_name = $1
		`
	case *dialect.MySQL:
		query = `
			SELECT column_name
			FROM information_schema.columns
			WHERE table_name = ?
			AND table_schema = DATABASE()
		`
	case *dialect.SQLite:
		query = fmt.Sprintf("PRAGMA table_info(%s)", m.dialect.Quote(tableName))
	default:
		return nil, fmt.Errorf("dialeto não suportado")
	}
	
	rows, err := m.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	columns := make(map[string]bool)
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		columns[columnName] = true
	}
	
	return columns, rows.Err()
}

// recordMigration registra uma migração aplicada
func (m *Migrator) recordMigration(ctx context.Context, name string) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (name, timestamp, applied)
		VALUES (?, ?, ?)
	`, m.dialect.Quote("migrations"))
	
	_, err := m.db.ExecContext(ctx, query, name, time.Now(), true)
	return err
}

// GetAppliedMigrations retorna todas as migrações aplicadas
func (m *Migrator) GetAppliedMigrations(ctx context.Context) ([]Migration, error) {
	query := fmt.Sprintf(`
		SELECT id, name, timestamp, applied
		FROM %s
		ORDER BY timestamp
	`, m.dialect.Quote("migrations"))
	
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var migrations []Migration
	for rows.Next() {
		var migration Migration
		if err := rows.Scan(
			&migration.ID,
			&migration.Name,
			&migration.Timestamp,
			&migration.Applied,
		); err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}
	
	return migrations, rows.Err()
}
