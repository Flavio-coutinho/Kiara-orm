package hooks

import (
	"context"
	"reflect"
)

// HookType representa os tipos de hooks disponíveis
type HookType int

const (
	BeforeCreate HookType = iota
	AfterCreate
	BeforeUpdate
	AfterUpdate
	BeforeDelete
	AfterDelete
	BeforeQuery
	AfterQuery
)

// Hook representa uma função de hook
type Hook func(ctx context.Context, value interface{}) error

// HookManager gerencia os hooks do ORM
type HookManager struct {
	hooks map[HookType][]Hook
}

// NewHookManager cria uma nova instância do HookManager
func NewHookManager() *HookManager {
	return &HookManager{
		hooks: make(map[HookType][]Hook),
	}
}

// Register registra um novo hook
func (hm *HookManager) Register(hookType HookType, hook Hook) {
	if hm.hooks[hookType] == nil {
		hm.hooks[hookType] = make([]Hook, 0)
	}
	hm.hooks[hookType] = append(hm.hooks[hookType], hook)
}

// Execute executa todos os hooks de um determinado tipo
func (hm *HookManager) Execute(ctx context.Context, hookType HookType, value interface{}) error {
	hooks := hm.hooks[hookType]
	for _, hook := range hooks {
		if err := hook(ctx, value); err != nil {
			return err
		}
	}
	return nil
}

// HasHooks verifica se existem hooks para um tipo específico
func (hm *HookManager) HasHooks(hookType HookType) bool {
	return len(hm.hooks[hookType]) > 0
} 