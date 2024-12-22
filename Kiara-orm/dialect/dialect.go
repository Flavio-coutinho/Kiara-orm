package dialect

import "github.com/Flavio-coutinho/Kiara-orm/types"

// Dialect define a interface que todos os dialetos SQL devem implementar
type Dialect interface {
    // GetDataTypeSQL converte um tipo do ORM para o tipo SQL correspondente
    GetDataTypeSQL(field types.FieldMapping) string
    
    // Quote coloca identificadores entre aspas de acordo com o dialeto
    Quote(identifier string) string
    
    // Placeholder retorna o placeholder para parâmetros preparados (?, $1, etc)
    Placeholder(index int) string
    
    // AutoIncrementSQL retorna a sintaxe para auto incremento
    AutoIncrementSQL() string
    
    // CreateTableSQL gera o SQL para criar uma tabela
    CreateTableSQL(table types.TableMapping) string
    
    // AddColumnSQL gera o SQL para adicionar uma coluna
    AddColumnSQL(table string, field types.FieldMapping) string
    
    // DropColumnSQL gera o SQL para remover uma coluna
    DropColumnSQL(table, column string) string
    
    // CreateIndexSQL gera o SQL para criar um índice
    CreateIndexSQL(table, indexName string, columns []string, unique bool) string
} 