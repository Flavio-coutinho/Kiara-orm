package dialect

import (
	"fmt"
	"strings"
	
	"github.com/Flavio-coutinho/Kiara-orm/types"
)

type SQLite struct{}

func NewSQLite() *SQLite {
	return &SQLite{}
}

func (s *SQLite) GetDataTypeSQL(field types.FieldMapping) string {
	switch field.Type {
	case types.Integer:
		return "INTEGER"
	case types.Float:
		return "REAL"
	case types.Text:
		if field.Size > 0 {
			return fmt.Sprintf("VARCHAR(%d)", field.Size)
		}
		return "TEXT"
	case types.Boolean:
		return "BOOLEAN"
	case types.DateTime:
		return "DATETIME"
	case types.Date:
		return "DATE"
	case types.Time:
		return "TIME"
	default:
		return "TEXT"
	}
}

func (s *SQLite) Quote(identifier string) string {
	return `"` + identifier + `"`
}

func (s *SQLite) Placeholder(_ int) string {
	return "?"
}

func (s *SQLite) AutoIncrementSQL() string {
	return "AUTOINCREMENT"
}

func (s *SQLite) CreateTableSQL(table types.TableMapping) string {
	var builder strings.Builder
	
	builder.WriteString("CREATE TABLE IF NOT EXISTS ")
	builder.WriteString(s.Quote(table.TableName))
	builder.WriteString(" (\n")
	
	var columns []string
	var primaryKeys []string
	
	for _, field := range table.Fields {
		column := fmt.Sprintf("  %s %s", s.Quote(field.Name), s.GetDataTypeSQL(field))
		
		if !field.IsNullable {
			column += " NOT NULL"
		}
		
		if field.IsAutoInc {
			if field.IsPrimaryKey {
				column += " PRIMARY KEY"
			}
			column += " " + s.AutoIncrementSQL()
		}
		
		if field.IsUnique {
			column += " UNIQUE"
		}
		
		if field.IsPrimaryKey && !field.IsAutoInc {
			primaryKeys = append(primaryKeys, s.Quote(field.Name))
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

func (s *SQLite) AddColumnSQL(table string, field types.FieldMapping) string {
	var builder strings.Builder
	
	builder.WriteString("ALTER TABLE ")
	builder.WriteString(s.Quote(table))
	builder.WriteString(" ADD COLUMN ")
	builder.WriteString(s.Quote(field.Name))
	builder.WriteString(" ")
	builder.WriteString(s.GetDataTypeSQL(field))
	
	if !field.IsNullable {
		builder.WriteString(" NOT NULL")
	}
	
	if field.IsUnique {
		builder.WriteString(" UNIQUE")
	}
	
	return builder.String()
}

func (s *SQLite) DropColumnSQL(table, column string) string {
	// SQLite não suporta DROP COLUMN diretamente
	// É necessário recriar a tabela sem a coluna
	return fmt.Sprintf("-- SQLite não suporta DROP COLUMN diretamente.\n"+
		"-- É necessário recriar a tabela sem a coluna %s", s.Quote(column))
}

func (s *SQLite) CreateIndexSQL(table, indexName string, columns []string, unique bool) string {
	var builder strings.Builder
	
	builder.WriteString("CREATE ")
	if unique {
		builder.WriteString("UNIQUE ")
	}
	builder.WriteString("INDEX ")
	builder.WriteString(s.Quote(indexName))
	builder.WriteString(" ON ")
	builder.WriteString(s.Quote(table))
	builder.WriteString(" (")
	
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = s.Quote(col)
	}
	
	builder.WriteString(strings.Join(quotedColumns, ", "))
	builder.WriteString(")")
	
	return builder.String()
}
