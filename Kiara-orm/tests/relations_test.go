package tests

import (
	"context"
	"testing"
	
	"github.com/Flavio-coutinho/kiara-orm/query"
	"github.com/Flavio-coutinho/kiara-orm/tests/models"
)

func TestRelations(t *testing.T) {
	sess := setupTestSession(t)
	ctx := context.Background()
	
	// Setup relationships
	sess.HasOne(&models.Post{}, "User", &models.User{}, "UserID")
	
	// Create test data
	user := &models.User{
		Name:  "Jane Doe",
		Email: "jane@example.com",
		Age:   30,
	}
	
	err := sess.Model(&models.User{}).Create(ctx, user)
	if err != nil {
		t.Fatalf("Falha ao criar usuário: %v", err)
	}
	
	post := &models.Post{
		Title:   "Test Post",
		Content: "Test Content",
		UserID:  user.ID,
	}
	
	err = sess.Model(&models.Post{}).Create(ctx, post)
	if err != nil {
		t.Fatalf("Falha ao criar post: %v", err)
	}
	
	// Test preload
	t.Run("Preload", func(t *testing.T) {
		var loadedPost models.Post
		err := sess.Model(&models.Post{}).
			Preload("User").
			Find(ctx, &loadedPost,
				query.Condition{Column: "id", Operation: query.OpEq, Value: post.ID})
		
		if err != nil {
			t.Errorf("Falha ao carregar post com usuário: %v", err)
		}
		
		if loadedPost.User == nil {
			t.Error("Usuário não foi carregado")
		}
		
		if loadedPost.User.Name != "Jane Doe" {
			t.Errorf("Nome do usuário esperado 'Jane Doe', recebido '%s'", loadedPost.User.Name)
		}
	})
} 