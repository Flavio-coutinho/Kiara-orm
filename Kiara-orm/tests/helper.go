package tests

import (
	"testing"
	
	"github.com/Flavio-coutinho/kiara-orm/connection"
	"github.com/Flavio-coutinho/kiara-orm/session"
	"github.com/Flavio-coutinho/kiara-orm/tests/config"
)

func setupTestSession(t *testing.T) *session.Session {
	cfg := config.TestConfig()
	pool, err := connection.NewPool(cfg)
	if err != nil {
		t.Fatalf("Falha ao criar pool: %v", err)
	}
	
	sess := session.NewSession(pool.GetDB(), config.TestDialect())
	
	// Clean up
	t.Cleanup(func() {
		pool.Close()
	})
	
	return sess
} 