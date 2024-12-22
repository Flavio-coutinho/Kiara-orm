package types

// DataType representa os tipos de dados suportados pelo ORM
type DataType int

const (
    Unknown DataType = iota
    Integer
    Float
    Text
    Boolean
    DateTime
    Date
    Time
)

// FieldMapping representa o mapeamento de um campo da struct para o banco de dados
type FieldMapping struct {
    Name         string
    Type         DataType
    Size         int
    IsPrimaryKey bool
    IsAutoInc    bool
    IsNullable   bool
    IsUnique     bool
}

// TableMapping representa o mapeamento de uma struct para uma tabela
type TableMapping struct {
    TableName string
    Fields    []FieldMapping
}

// TypeMapper é responsável por converter tipos Go para tipos do banco de dados
type TypeMapper struct {
    mappings map[string]DataType
}

// NewTypeMapper cria uma nova instância do TypeMapper
func NewTypeMapper() *TypeMapper {
    m := &TypeMapper{
        mappings: make(map[string]DataType),
    }
    m.initDefaultMappings()
    return m
}

// initDefaultMappings inicializa os mapeamentos padrão de tipos Go para tipos do ORM
func (tm *TypeMapper) initDefaultMappings() {
    tm.mappings["string"] = Text
    tm.mappings["int"] = Integer
    tm.mappings["int64"] = Integer
    tm.mappings["float64"] = Float
    tm.mappings["bool"] = Boolean
    tm.mappings["time.Time"] = DateTime
}

// GetDataType retorna o tipo de dados do ORM correspondente ao tipo Go
func (tm *TypeMapper) GetDataType(goType string) DataType {
    if dt, ok := tm.mappings[goType]; ok {
        return dt
    }
    return Unknown
}
