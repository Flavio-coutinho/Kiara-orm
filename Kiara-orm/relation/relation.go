package relation

import (
	"fmt"
	"reflect"
)

// RelationType representa o tipo de relacionamento
type RelationType int

const (
	OneToOne RelationType = iota
	OneToMany
	ManyToMany
)

// Relation representa um relacionamento entre modelos
type Relation struct {
	Type         RelationType
	Model        interface{}
	ForeignKey   string
	ReferenceKey string
	JoinTable    string // Para Many-to-Many
	Preload      bool
}

// RelationManager gerencia os relacionamentos entre modelos
type RelationManager struct {
	relations map[string]map[string]Relation // map[model][field]Relation
}

// NewRelationManager cria uma nova instância do RelationManager
func NewRelationManager() *RelationManager {
	return &RelationManager{
		relations: make(map[string]map[string]Relation),
	}
}

// HasOne define um relacionamento um-para-um
func (rm *RelationManager) HasOne(model interface{}, field string, related interface{}, foreignKey string) {
	rm.addRelation(model, field, Relation{
		Type:         OneToOne,
		Model:        related,
		ForeignKey:   foreignKey,
		ReferenceKey: "ID", // Assume ID como chave padrão
		Preload:      false,
	})
}

// HasMany define um relacionamento um-para-muitos
func (rm *RelationManager) HasMany(model interface{}, field string, related interface{}, foreignKey string) {
	rm.addRelation(model, field, Relation{
		Type:         OneToMany,
		Model:        related,
		ForeignKey:   foreignKey,
		ReferenceKey: "ID",
		Preload:      false,
	})
}

// ManyToMany define um relacionamento muitos-para-muitos
func (rm *RelationManager) ManyToMany(model interface{}, field string, related interface{}, joinTable string) {
	rm.addRelation(model, field, Relation{
		Type:         ManyToMany,
		Model:        related,
		JoinTable:    joinTable,
		ReferenceKey: "ID",
		Preload:      false,
	})
}

// EnablePreload habilita o carregamento automático de um relacionamento
func (rm *RelationManager) EnablePreload(model interface{}, field string) {
	modelName := rm.getModelName(model)
	if relations, ok := rm.relations[modelName]; ok {
		if relation, ok := relations[field]; ok {
			relation.Preload = true
			relations[field] = relation
		}
	}
}

// GetRelation retorna um relacionamento específico
func (rm *RelationManager) GetRelation(model interface{}, field string) (Relation, bool) {
	modelName := rm.getModelName(model)
	if relations, ok := rm.relations[modelName]; ok {
		relation, ok := relations[field]
		return relation, ok
	}
	return Relation{}, false
}

// GetPreloadFields retorna todos os campos que devem ser pré-carregados
func (rm *RelationManager) GetPreloadFields(model interface{}) []string {
	modelName := rm.getModelName(model)
	fields := make([]string, 0)
	
	if relations, ok := rm.relations[modelName]; ok {
		for field, relation := range relations {
			if relation.Preload {
				fields = append(fields, field)
			}
		}
	}
	
	return fields
}

// addRelation adiciona um relacionamento ao gerenciador
func (rm *RelationManager) addRelation(model interface{}, field string, relation Relation) {
	modelName := rm.getModelName(model)
	if rm.relations[modelName] == nil {
		rm.relations[modelName] = make(map[string]Relation)
	}
	rm.relations[modelName][field] = relation
}

// getModelName retorna o nome do modelo
func (rm *RelationManager) getModelName(model interface{}) string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
} 