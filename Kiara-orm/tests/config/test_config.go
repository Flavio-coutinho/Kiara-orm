package config

import (
	"github.com/Flavio-coutinho/kiara-orm/connection"
	"github.com/Flavio-coutinho/kiara-orm/dialect"
)

// TestConfig retorna uma configuração para testes
func TestConfig() *connection.Config {
	return &connection.Config{
		Type:     connection.MySQL,
		Host:     "localhost",
		Port:     3306,
		User:     "test_user",
		Password: "test_pass",
		Database: "kiara_test",
	}
}

// TestDialect retorna um dialeto para testes
func TestDialect() dialect.Dialect {
	return dialect.NewMySQL()
} 