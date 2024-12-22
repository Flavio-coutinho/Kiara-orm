package dialect

import (
	"fmt"
	"strings"
	
	"github.com/Flavio-coutinho/Kiara-orm/types"
)

type PostgreSQL struct{}

func NewPostgreSQL() *PostgreSQL {
	return &PostgreSQL{}
}

func (p *PostgreSQL) GetDataTypeSQL(field types.FieldMapping) string {
	switch field.Type {
	case types.Integer:
		if field.IsAutoInc {
			return "SERIAL"
		}
		return "INTEGER"
	case types.Float:
		return "DOUBLE PRECISION"
	case types.Text:
		if field.Size > 0 {
			return fmt.Sprintf("VARCHAR(%d)", field.Size)
		}
		return "TEXT"
	case types.Boolean:
		return "BOOLEAN"
	case types.DateTime:
		return "TIMESTAMP"
	case types.Date:
		return "DATE"
	case types.Time:
		return "TIME"
	default:
		return "TEXT"
	}
}

func (p *PostgreSQL) Quote(identifier string) string {
	return `"` + identifier + `"`
}

func (p *PostgreSQL) Placeholder(index int) string {
	return fmt.Sprintf("$%d", index)
}

func (p *PostgreSQL) AutoIncrementSQL() string {
	return "" // Não necessário no PostgreSQL, usa SERIAL
}

func (p *PostgreSQL) CreateTableSQL(table types.TableMapping) string {
	var builder strings.Builder
	
	builder.WriteString("CREATE TABLE IF NOT EXISTS ")
	builder.WriteString(p.Quote(table.TableName))
	builder.WriteString(" (\n")
	
	var columns []string
	var primaryKeys []string
	
	for _, field := range table.Fields {
		column := fmt.Sprintf("  %s %s", p.Quote(field.Name), p.GetDataTypeSQL(field))
		
		if !field.IsNullable {
			column += " NOT NULL"
		}
		
		if field.IsUnique {
			column += " UNIQUE"
		}
		
		if field.IsPrimaryKey {
			primaryKeys = append(primaryKeys, p.Quote(field.Name))
		}
		
		columns = append(columns, column)
	}
	
	if len(primaryKeys) > 0 {
		columns = append(columns, fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}
	
	builder.WriteString(strings.Join(columns, ",\n"))
	builder.WriteString("\n);")
	
	return builder.String()
}

func (p *PostgreSQL) AddColumnSQL(table string, field types.FieldMapping) string {
	var builder strings.Builder
	
	builder.WriteString("ALTER TABLE ")
	builder.WriteString(p.Quote(table))
	builder.WriteString(" ADD COLUMN ")
	builder.WriteString(p.Quote(field.Name))
	builder.WriteString(" ")
	builder.WriteString(p.GetDataTypeSQL(field))
	
	if !field.IsNullable {
		builder.WriteString(" NOT NULL")
	}
	
	if field.IsUnique {
		builder.WriteString(" UNIQUE")
	}
	
	return builder.String()
}

func (p *PostgreSQL) DropColumnSQL(table, column string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s",
		p.Quote(table),
		p.Quote(column))
}

func (p *PostgreSQL) CreateIndexSQL(table, indexName string, columns []string, unique bool) string {
	var builder strings.Builder
	
	builder.WriteString("CREATE ")
	if unique {
		builder.WriteString("UNIQUE ")
	}
	builder.WriteString("INDEX ")
	builder.WriteString(p.Quote(indexName))
	builder.WriteString(" ON ")
	builder.WriteString(p.Quote(table))
	builder.WriteString(" (")
	
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = p.Quote(col)
	}
	
	builder.WriteString(strings.Join(quotedColumns, ", "))
	builder.WriteString(")")
	
	return builder.String()
}
