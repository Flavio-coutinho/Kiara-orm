package tests

import (
	"context"
	"testing"
	"time"
	
	"github.com/Flavio-coutinho/kiara-orm/session"
	"github.com/Flavio-coutinho/kiara-orm/tests/config"
	"github.com/Flavio-coutinho/kiara-orm/tests/models"
)

func TestCRUD(t *testing.T) {
	// Setup
	cfg := config.TestConfig()
	pool, err := connection.NewPool(cfg)
	if err != nil {
		t.Fatalf("Falha ao criar pool: %v", err)
	}
	defer pool.Close()
	
	sess := session.NewSession(pool.GetDB(), config.TestDialect())
	ctx := context.Background()
	
	// Migrate
	err = sess.AutoMigrate(ctx, &models.User{}, &models.Post{})
	if err != nil {
		t.Fatalf("Falha na migração: %v", err)
	}
	
	// Test Create
	t.Run("Create", func(t *testing.T) {
		user := &models.User{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   25,
		}
		
		err := sess.Model(&models.User{}).Create(ctx, user)
		if err != nil {
			t.Errorf("Falha ao criar usuário: %v", err)
		}
		
		if user.ID == 0 {
			t.Error("ID não foi definido após criação")
		}
	})
	
	// Test Read
	t.Run("Read", func(t *testing.T) {
		var user models.User
		err := sess.Model(&models.User{}).Find(ctx, &user,
			query.Condition{Column: "email", Operation: query.OpEq, Value: "john@example.com"})
		
		if err != nil {
			t.Errorf("Falha ao buscar usuário: %v", err)
		}
		
		if user.Name != "John Doe" {
			t.Errorf("Nome esperado 'John Doe', recebido '%s'", user.Name)
		}
	})
	
	// Test Update
	t.Run("Update", func(t *testing.T) {
		updates := &models.User{
			Name: "John Updated",
		}
		
		err := sess.Model(&models.User{}).Update(ctx, updates,
			query.Condition{Column: "email", Operation: query.OpEq, Value: "john@example.com"})
		
		if err != nil {
			t.Errorf("Falha ao atualizar usuário: %v", err)
		}
		
		// Verify update
		var user models.User
		err = sess.Model(&models.User{}).Find(ctx, &user,
			query.Condition{Column: "email", Operation: query.OpEq, Value: "john@example.com"})
		
		if user.Name != "John Updated" {
			t.Errorf("Nome esperado 'John Updated', recebido '%s'", user.Name)
		}
	})
	
	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		err := sess.Model(&models.User{}).Delete(ctx,
			query.Condition{Column: "email", Operation: query.OpEq, Value: "john@example.com"})
		
		if err != nil {
			t.Errorf("Falha ao deletar usuário: %v", err)
		}
		
		// Verify deletion
		var user models.User
		err = sess.Model(&models.User{}).Find(ctx, &user,
			query.Condition{Column: "email", Operation: query.OpEq, Value: "john@example.com"})
		
		if err == nil {
			t.Error("Usuário ainda existe após deleção")
		}
	})
} 