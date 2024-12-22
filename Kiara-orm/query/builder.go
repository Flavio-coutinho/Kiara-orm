package query

import (
	"fmt"
	"strings"
	
	"github.com/Flavio-coutinho/Kiara-orm/dialect"
)

// Operation representa os tipos de operações de comparação
type Operation string

const (
	OpEq    Operation = "="
	OpNe    Operation = "<>"
	OpGt    Operation = ">"
	OpGe    Operation = ">="
	OpLt    Operation = "<"
	OpLe    Operation = "<="
	OpLike  Operation = "LIKE"
	OpILike Operation = "ILIKE"
	OpIn    Operation = "IN"
)

// Condition representa uma condição WHERE
type Condition struct {
	Column    string
	Operation Operation
	Value     interface{}
}

// Builder é responsável por construir queries SQL
type Builder struct {
	dialect    dialect.Dialect
	table      string
	columns    []string
	conditions []Condition
	orderBy    []string
	limit      *int
	offset     *int
	params     []interface{}
	joins      []string
	groupBy    []string
	having     []Condition
}

// NewBuilder cria uma nova instância do Builder
func NewBuilder(dialect dialect.Dialect) *Builder {
	return &Builder{
		dialect:    dialect,
		columns:    make([]string, 0),
		conditions: make([]Condition, 0),
		params:     make([]interface{}, 0),
	}
}

// Table define a tabela para a query
func (b *Builder) Table(table string) *Builder {
	b.table = table
	return b
}

// Select define as colunas para selecionar
func (b *Builder) Select(columns ...string) *Builder {
	b.columns = append(b.columns, columns...)
	return b
}

// Where adiciona uma condição WHERE
func (b *Builder) Where(column string, op Operation, value interface{}) *Builder {
	b.conditions = append(b.conditions, Condition{
		Column:    column,
		Operation: op,
		Value:     value,
	})
	b.params = append(b.params, value)
	return b
}

// OrderBy adiciona ordenação
func (b *Builder) OrderBy(column string, desc bool) *Builder {
	order := b.dialect.Quote(column)
	if desc {
		order += " DESC"
	}
	b.orderBy = append(b.orderBy, order)
	return b
}

// Limit define o limite de resultados
func (b *Builder) Limit(limit int) *Builder {
	b.limit = &limit
	return b
}

// Offset define o offset dos resultados
func (b *Builder) Offset(offset int) *Builder {
	b.offset = &offset
	return b
}

// Join adiciona uma cláusula JOIN
func (b *Builder) Join(joinType, table, condition string) *Builder {
	join := fmt.Sprintf("%s JOIN %s ON %s", 
		joinType,
		b.dialect.Quote(table),
		condition)
	b.joins = append(b.joins, join)
	return b
}

// GroupBy adiciona agrupamento
func (b *Builder) GroupBy(columns ...string) *Builder {
	for _, col := range columns {
		b.groupBy = append(b.groupBy, b.dialect.Quote(col))
	}
	return b
}

// Having adiciona uma condição HAVING
func (b *Builder) Having(column string, op Operation, value interface{}) *Builder {
	b.having = append(b.having, Condition{
		Column:    column,
		Operation: op,
		Value:     value,
	})
	b.params = append(b.params, value)
	return b
}

// BuildSelect constrói uma query SELECT
func (b *Builder) BuildSelect() (string, []interface{}) {
	var builder strings.Builder
	
	builder.WriteString("SELECT ")
	
	// Colunas
	if len(b.columns) == 0 {
		builder.WriteString("*")
	} else {
		quotedColumns := make([]string, len(b.columns))
		for i, col := range b.columns {
			quotedColumns[i] = b.dialect.Quote(col)
		}
		builder.WriteString(strings.Join(quotedColumns, ", "))
	}
	
	// FROM
	builder.WriteString(" FROM ")
	builder.WriteString(b.dialect.Quote(b.table))
	
	// JOINs
	if len(b.joins) > 0 {
		builder.WriteString(" ")
		builder.WriteString(strings.Join(b.joins, " "))
	}
	
	// WHERE
	if len(b.conditions) > 0 {
		builder.WriteString(" WHERE ")
		whereConditions := make([]string, len(b.conditions))
		for i, cond := range b.conditions {
			whereConditions[i] = fmt.Sprintf("%s %s %s",
				b.dialect.Quote(cond.Column),
				cond.Operation,
				b.dialect.Placeholder(i+1))
		}
		builder.WriteString(strings.Join(whereConditions, " AND "))
	}
	
	// GROUP BY
	if len(b.groupBy) > 0 {
		builder.WriteString(" GROUP BY ")
		builder.WriteString(strings.Join(b.groupBy, ", "))
	}
	
	// HAVING
	if len(b.having) > 0 {
		builder.WriteString(" HAVING ")
		havingConditions := make([]string, len(b.having))
		paramOffset := len(b.conditions)
		for i, cond := range b.having {
			havingConditions[i] = fmt.Sprintf("%s %s %s",
				b.dialect.Quote(cond.Column),
				cond.Operation,
				b.dialect.Placeholder(paramOffset+i+1))
		}
		builder.WriteString(strings.Join(havingConditions, " AND "))
	}
	
	// ORDER BY
	if len(b.orderBy) > 0 {
		builder.WriteString(" ORDER BY ")
		builder.WriteString(strings.Join(b.orderBy, ", "))
	}
	
	// LIMIT
	if b.limit != nil {
		builder.WriteString(fmt.Sprintf(" LIMIT %d", *b.limit))
	}
	
	// OFFSET
	if b.offset != nil {
		builder.WriteString(fmt.Sprintf(" OFFSET %d", *b.offset))
	}
	
	return builder.String(), b.params
}
