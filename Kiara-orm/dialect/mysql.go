package dialect

import (
	"fmt"
	"strings"
	
	"github.com/Flavio-coutinho/Kiara-orm/types"
)

type MySQL struct{}

func NewMySQL() *MySQL {
	return &MySQL{}
}

func (m *MySQL) GetDataTypeSQL(field types.FieldMapping) string {
	switch field.Type {
	case types.Integer:
		if field.IsAutoInc {
			return "INT"
		}
		return "INTEGER"
	case types.Float:
		return "DOUBLE"
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

func (m *MySQL) Quote(identifier string) string {
	return "`" + identifier + "`"
}

func (m *MySQL) Placeholder(_ int) string {
	return "?"
}

func (m *MySQL) AutoIncrementSQL() string {
	return "AUTO_INCREMENT"
}

func (m *MySQL) CreateTableSQL(table types.TableMapping) string {
	var builder strings.Builder
	
	builder.WriteString("CREATE TABLE IF NOT EXISTS ")
	builder.WriteString(m.Quote(table.TableName))
	builder.WriteString(" (\n")
	
	var columns []string
	var primaryKeys []string
	
	for _, field := range table.Fields {
		column := fmt.Sprintf("  %s %s", m.Quote(field.Name), m.GetDataTypeSQL(field))
		
		if !field.IsNullable {
			column += " NOT NULL"
		}
		
		if field.IsAutoInc {
			column += " " + m.AutoIncrementSQL()
		}
		
		if field.IsUnique {
			column += " UNIQUE"
		}
		
		if field.IsPrimaryKey {
			primaryKeys = append(primaryKeys, m.Quote(field.Name))
		}
		
		columns = append(columns, column)
	}
	
	if len(primaryKeys) > 0 {
		columns = append(columns, fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}
	
	builder.WriteString(strings.Join(columns, ",\n"))
	builder.WriteString("\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;")
	
	return builder.String()
}

func (m *MySQL) AddColumnSQL(table string, field types.FieldMapping) string {
	var builder strings.Builder
	
	builder.WriteString("ALTER TABLE ")
	builder.WriteString(m.Quote(table))
	builder.WriteString(" ADD COLUMN ")
	builder.WriteString(m.Quote(field.Name))
	builder.WriteString(" ")
	builder.WriteString(m.GetDataTypeSQL(field))
	
	if !field.IsNullable {
		builder.WriteString(" NOT NULL")
	}
	
	if field.IsUnique {
		builder.WriteString(" UNIQUE")
	}
	
	return builder.String()
}

func (m *MySQL) DropColumnSQL(table, column string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s",
		m.Quote(table),
		m.Quote(column))
}

func (m *MySQL) CreateIndexSQL(table, indexName string, columns []string, unique bool) string {
	var builder strings.Builder
	
	builder.WriteString("CREATE ")
	if unique {
		builder.WriteString("UNIQUE ")
	}
	builder.WriteString("INDEX ")
	builder.WriteString(m.Quote(indexName))
	builder.WriteString(" ON ")
	builder.WriteString(m.Quote(table))
	builder.WriteString(" (")
	
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = m.Quote(col)
	}
	
	builder.WriteString(strings.Join(quotedColumns, ", "))
	builder.WriteString(")")
	
	return builder.String()
}
