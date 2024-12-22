package scope

import (
	"context"
	"github.com/Flavio-coutinho/Kiara-orm/query"
)

// Scope define uma função que modifica uma query
type Scope func(ctx context.Context, builder *query.Builder) *query.Builder

// ScopeManager gerencia os scopes registrados
type ScopeManager struct {
	globalScopes map[string]Scope
	modelScopes  map[string]map[string]Scope // map[model][name]Scope
}

// NewScopeManager cria uma nova instância do ScopeManager
func NewScopeManager() *ScopeManager {
	return &ScopeManager{
		globalScopes: make(map[string]Scope),
		modelScopes:  make(map[string]map[string]Scope),
	}
}

// AddGlobalScope adiciona um scope global
func (sm *ScopeManager) AddGlobalScope(name string, scope Scope) {
	sm.globalScopes[name] = scope
}

// AddModelScope adiciona um scope específico para um modelo
func (sm *ScopeManager) AddModelScope(model interface{}, name string, scope Scope) {
	modelName := getModelName(model)
	if sm.modelScopes[modelName] == nil {
		sm.modelScopes[modelName] = make(map[string]Scope)
	}
	sm.modelScopes[modelName][name] = scope
}

// ApplyScopes aplica todos os scopes ativos a uma query
func (sm *ScopeManager) ApplyScopes(ctx context.Context, model interface{}, builder *query.Builder) *query.Builder {
	// Aplica scopes globais
	for _, scope := range sm.globalScopes {
		builder = scope(ctx, builder)
	}
	
	// Aplica scopes do modelo
	modelName := getModelName(model)
	if scopes, ok := sm.modelScopes[modelName]; ok {
		for _, scope := range scopes {
			builder = scope(ctx, builder)
		}
	}
	
	return builder
} 