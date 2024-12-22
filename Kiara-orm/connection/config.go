package connection

import "time"

// DBType representa o tipo de banco de dados suportado
type DBType string

const (
    MySQL    DBType = "mysql"
    Postgres DBType = "postgres"
    SQLite   DBType = "sqlite"
)

// Config representa as configurações de conexão com o banco de dados
type Config struct {
    Type     DBType
    Host     string
    Port     int
    User     string
    Password string
    Database string
    
    // Configurações do pool de conexões
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    
    // Configurações específicas do SQLite
    SQLitePath string
}

// NewConfig cria uma nova configuração com valores padrão
func NewConfig() *Config {
    return &Config{
        Type:           MySQL,
        Host:           "localhost",
        Port:           3306,
        MaxOpenConns:   10,
        MaxIdleConns:   5,
        ConnMaxLifetime: time.Hour,
    }
}

// DSN retorna a string de conexão baseada no tipo de banco de dados
func (c *Config) DSN() string {
    switch c.Type {
    case MySQL:
        return c.mysqlDSN()
    case Postgres:
        return c.postgresDSN()
    case SQLite:
        return c.sqliteDSN()
    default:
        return ""
    }
}

func (c *Config) mysqlDSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", 
        c.User, c.Password, c.Host, c.Port, c.Database)
}

func (c *Config) postgresDSN() string {
    return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        c.Host, c.Port, c.User, c.Password, c.Database)
}

func (c *Config) sqliteDSN() string {
    return c.SQLitePath
}
