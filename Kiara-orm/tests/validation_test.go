package tests

import (
	"context"
	"testing"
	
	"github.com/Flavio-coutinho/kiara-orm/tests/models"
)

func TestValidation(t *testing.T) {
	sess := setupTestSession(t)
	ctx := context.Background()
	
	t.Run("Required Field", func(t *testing.T) {
		user := &models.User{
			Email: "test@example.com",
			Age:   20,
		}
		
		err := sess.Model(&models.User{}).Create(ctx, user)
		if err == nil {
			t.Error("Validação deveria falhar para nome vazio")
		}
	})
	
	t.Run("Email Format", func(t *testing.T) {
		user := &models.User{
			Name:  "Test User",
			Email: "invalid-email",
			Age:   20,
		}
		
		err := sess.Model(&models.User{}).Create(ctx, user)
		if err == nil {
			t.Error("Validação deveria falhar para email inválido")
		}
	})
	
	t.Run("Minimum Age", func(t *testing.T) {
		user := &models.User{
			Name:  "Test User",
			Email: "test@example.com",
			Age:   16,
		}
		
		err := sess.Model(&models.User{}).Create(ctx, user)
		if err == nil {
			t.Error("Validação deveria falhar para idade menor que 18")
		}
	})
} 