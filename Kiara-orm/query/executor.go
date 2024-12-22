package query

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	
	"github.com/Flavio-coutinho/Kiara-orm/types"
)

// Executor é responsável por executar queries SQL
type Executor struct {
	db      *sql.DB
	builder *Builder
}

// NewExecutor cria uma nova instância do Executor
func NewExecutor(db *sql.DB, builder *Builder) *Executor {
	return &Executor{
		db:      db,
		builder: builder,
	}
}

// QueryRow executa uma query e retorna uma única linha
func (e *Executor) QueryRow(ctx context.Context, dest interface{}) error {
	query, params := e.builder.BuildSelect()
	
	row := e.db.QueryRowContext(ctx, query, params...)
	return e.scanRow(row, dest)
}

// Query executa uma query e retorna múltiplas linhas
func (e *Executor) Query(ctx context.Context, dest interface{}) error {
	query, params := e.builder.BuildSelect()
	
	rows, err := e.db.QueryContext(ctx, query, params...)
	if err != nil {
		return fmt.Errorf("erro ao executar query: %v", err)
	}
	defer rows.Close()
	
	return e.scanRows(rows, dest)
}

// scanRow faz o scan de uma única linha para uma struct
func (e *Executor) scanRow(row *sql.Row, dest interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("destino deve ser um ponteiro")
	}
	
	v = v.Elem()
	t := v.Type()
	
	// Cria slice para armazenar os endereços dos campos
	values := make([]interface{}, len(e.builder.columns))
	for i, col := range e.builder.columns {
		field, ok := t.FieldByName(col)
		if !ok {
			return fmt.Errorf("campo %s não encontrado na struct", col)
		}
		
		values[i] = v.FieldByName(field.Name).Addr().Interface()
	}
	
	return row.Scan(values...)
}

// scanRows faz o scan de múltiplas linhas para um slice de structs
func (e *Executor) scanRows(rows *sql.Rows, dest interface{}) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("destino deve ser um ponteiro para slice")
	}
	
	sliceVal := v.Elem()
	elemType := sliceVal.Type().Elem()
	
	for rows.Next() {
		// Cria nova instância do tipo do elemento
		elem := reflect.New(elemType).Elem()
		
		// Cria slice para armazenar os endereços dos campos
		values := make([]interface{}, len(e.builder.columns))
		for i, col := range e.builder.columns {
			field, ok := elemType.FieldByName(col)
			if !ok {
				return fmt.Errorf("campo %s não encontrado na struct", col)
			}
			
			values[i] = elem.FieldByName(field.Name).Addr().Interface()
		}
		
		if err := rows.Scan(values...); err != nil {
			return fmt.Errorf("erro ao fazer scan da linha: %v", err)
		}
		
		sliceVal.Set(reflect.Append(sliceVal, elem))
	}
	
	return rows.Err()
}
