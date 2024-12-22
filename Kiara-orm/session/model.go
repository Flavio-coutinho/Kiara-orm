package session

import (
	"context"
	"fmt"
	"reflect"
	"time"
	
	"github.com/Flavio-coutinho/kiara-orm/query"
	"github.com/Flavio-coutinho/kiara-orm/schema"
	"github.com/Flavio-coutinho/kiara-orm/hooks"
	"github.com/Flavio-coutinho/kiara-orm/logger"
	"github.com/Flavio-coutinho/kiara-orm/validator"
	"github.com/Flavio-coutinho/kiara-orm/softdelete"
	"github.com/Flavio-coutinho/kiara-orm/bulk"
	"github.com/Flavio-coutinho/kiara-orm/scope"
	"github.com/Flavio-coutinho/kiara-orm/pagination"
	"github.com/Flavio-coutinho/kiara-orm/metrics"
)

// ModelHandler manipula operações em um modelo específico
type ModelHandler struct {
	session *Session
	model   interface{}
	mapping *schema.TableMapping
	includeTrashed bool
	onlyTrashed bool
	preloadFields []string
	scopes    []scope.Scope
	paginator *pagination.Paginator
}

// NewModelHandler cria um novo manipulador de modelo
func NewModelHandler(session *Session, model interface{}) *ModelHandler {
	parser := schema.NewParser()
	mapping, _ := parser.Parse(model)
	
	return &ModelHandler{
		session: session,
		model:   model,
		mapping: mapping,
	}
}

// Create insere um novo registro
func (m *ModelHandler) Create(ctx context.Context, data interface{}) error {
	// Log da operação
	m.session.logger.Debug(ctx, "Iniciando criação de registro em %s", m.mapping.TableName)
	
	// Validação
	if err := m.session.validator.Validate(data); err != nil {
		m.session.logger.Error(ctx, "Validação falhou: %v", err)
		return err
	}
	
	// Executar hooks antes da criação
	if err := m.session.hooks.Execute(ctx, hooks.BeforeCreate, data); err != nil {
		return err
	}
	
	builder := m.session.Query().
		Table(m.mapping.TableName)
	
	// Extrai valores dos campos
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	columns := make([]string, 0)
	values := make([]interface{}, 0)
	
	for _, field := range m.mapping.Fields {
		if field.IsAutoInc {
			continue
		}
		
		columns = append(columns, field.Name)
		values = append(values, v.FieldByName(field.Name).Interface())
	}
	
	// Constrói e executa a query de inserção
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		m.session.dialect.Quote(m.mapping.TableName),
		m.buildColumnList(columns),
		m.buildPlaceholders(len(columns)),
	)
	
	_, err := m.session.db.ExecContext(ctx, query, values...)
	if err != nil {
		return err
	}
	
	// Executar hooks após a criação
	if err := m.session.hooks.Execute(ctx, hooks.AfterCreate, data); err != nil {
		return err
	}
	
	// Invalidar cache relacionado
	cacheKey := fmt.Sprintf("table:%s", m.mapping.TableName)
	m.session.cache.Delete(cacheKey)
	
	return nil
}

// Find busca registros
func (m *ModelHandler) Find(ctx context.Context, dest interface{}, conditions ...query.Condition) error {
	start := time.Now()
	
	builder := m.session.Query().Table(m.mapping.TableName)
	
	// Aplica scopes
	for _, scope := range m.scopes {
		builder = scope(ctx, builder)
	}
	
	// Aplica condições
	for _, cond := range conditions {
		builder.Where(cond.Column, cond.Operation, cond.Value)
	}
	
	// Aplica paginação
	if m.paginator != nil {
		// Primeiro, obtém o total de registros
		var count int64
		countBuilder := m.session.Query().
			Table(m.mapping.TableName).
			Select("COUNT(*) as count")
		
		err := m.session.Exec(countBuilder).QueryRow(ctx, &count)
		if err != nil {
			return err
		}
		
		m.paginator.SetTotal(count)
		
		// Aplica limit e offset
		builder.Limit(m.paginator.Limit()).
			Offset(m.paginator.Offset())
	}
	
	err := m.session.Exec(builder).Query(ctx, dest)
	
	// Registra métricas
	duration := time.Since(start).Seconds()
	m.session.metrics.AddMetric(metrics.QueryExecution, duration, map[string]string{
		"type":  "select",
		"table": m.mapping.TableName,
	})
	
	if err != nil {
		m.session.metrics.AddMetric(metrics.ErrorCount, 1, map[string]string{
			"type":      "query",
			"operation": "select",
			"table":     m.mapping.TableName,
		})
	}
	
	return err
}

// Update atualiza registros
func (m *ModelHandler) Update(ctx context.Context, data interface{}, conditions ...query.Condition) error {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	// Constrói o SET da query
	updates := make([]string, 0)
	values := make([]interface{}, 0)
	
	for _, field := range m.mapping.Fields {
		if field.IsPrimaryKey || field.IsAutoInc {
			continue
		}
		
		updates = append(updates, 
			fmt.Sprintf("%s = ?", m.session.dialect.Quote(field.Name)))
		values = append(values, v.FieldByName(field.Name).Interface())
	}
	
	// Adiciona condições WHERE
	where := make([]string, 0)
	for _, cond := range conditions {
		where = append(where,
			fmt.Sprintf("%s %s ?", 
				m.session.dialect.Quote(cond.Column),
				cond.Operation))
		values = append(values, cond.Value)
	}
	
	// Constrói e executa a query
	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		m.session.dialect.Quote(m.mapping.TableName),
		m.joinWithComma(updates),
		m.joinWithAnd(where),
	)
	
	_, err := m.session.db.ExecContext(ctx, query, values...)
	return err
}

// Delete remove registros
func (m *ModelHandler) Delete(ctx context.Context, conditions ...query.Condition) error {
	where := make([]string, 0)
	values := make([]interface{}, 0)
	
	for _, cond := range conditions {
		where = append(where,
			fmt.Sprintf("%s %s ?", 
				m.session.dialect.Quote(cond.Column),
				cond.Operation))
		values = append(values, cond.Value)
	}
	
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s",
		m.session.dialect.Quote(m.mapping.TableName),
		m.joinWithAnd(where),
	)
	
	_, err := m.session.db.ExecContext(ctx, query, values...)
	return err
}

// Funções auxiliares
func (m *ModelHandler) buildColumnList(columns []string) string {
	quoted := make([]string, len(columns))
	for i, col := range columns {
		quoted[i] = m.session.dialect.Quote(col)
	}
	return m.joinWithComma(quoted)
}

func (m *ModelHandler) buildPlaceholders(count int) string {
	placeholders := make([]string, count)
	for i := range placeholders {
		placeholders[i] = "?"
	}
	return m.joinWithComma(placeholders)
}

func (m *ModelHandler) joinWithComma(items []string) string {
	return m.joinWith(items, ", ")
}

func (m *ModelHandler) joinWithAnd(items []string) string {
	return m.joinWith(items, " AND ")
}

func (m *ModelHandler) joinWith(items []string, sep string) string {
	if len(items) == 0 {
		return ""
	}
	result := items[0]
	for _, item := range items[1:] {
		result += sep + item
	}
	return result
}

// Adicionar métodos auxiliares para cache
func (m *ModelHandler) buildCacheKey(conditions []query.Condition) string {
	// Criar uma chave única baseada na tabela e condições
	key := fmt.Sprintf("table:%s", m.mapping.TableName)
	for _, cond := range conditions {
		key += fmt.Sprintf("|%s:%v:%v", cond.Column, cond.Operation, cond.Value)
	}
	return key
}

func (m *ModelHandler) copyFromCache(cached, dest interface{}) error {
	// Copiar dados do cache para o destino usando reflection
	srcVal := reflect.ValueOf(cached)
	dstVal := reflect.ValueOf(dest)
	
	if dstVal.Kind() != reflect.Ptr {
		return fmt.Errorf("destino deve ser um ponteiro")
	}
	
	dstVal = dstVal.Elem()
	srcVal = reflect.Indirect(srcVal)
	
	if srcVal.Type() != dstVal.Type() {
		return fmt.Errorf("tipo do cache não corresponde ao tipo do destino")
	}
	
	dstVal.Set(srcVal)
	return nil
}

// SoftDelete realiza uma exclusão lógica
func (m *ModelHandler) SoftDelete(ctx context.Context, conditions ...query.Condition) error {
	now := time.Now()
	
	updates := map[string]interface{}{
		"deleted_at": &now,
	}
	
	return m.Update(ctx, updates, conditions...)
}

// Restore restaura registros excluídos logicamente
func (m *ModelHandler) Restore(ctx context.Context, conditions ...query.Condition) error {
	updates := map[string]interface{}{
		"deleted_at": nil,
	}
	
	return m.Update(ctx, updates, conditions...)
}

// WithTrashed inclui registros excluídos logicamente nas consultas
func (m *ModelHandler) WithTrashed() *ModelHandler {
	m.includeTrashed = true
	return m
}

// OnlyTrashed retorna apenas registros excluídos logicamente
func (m *ModelHandler) OnlyTrashed() *ModelHandler {
	m.onlyTrashed = true
	return m
}

// BulkCreate insere múltiplos registros
func (m *ModelHandler) BulkCreate(ctx context.Context, records []interface{}) error {
	bulkOp := bulk.NewBulkOperation(m.session.dialect, m.mapping, 1000)
	return bulkOp.BulkInsert(ctx, m.session.db, records)
}

// BulkUpdate atualiza múltiplos registros
func (m *ModelHandler) BulkUpdate(ctx context.Context, records []interface{}, conditions map[string]interface{}) error {
	bulkOp := bulk.NewBulkOperation(m.session.dialect, m.mapping, 1000)
	return bulkOp.BulkUpdate(ctx, m.session.db, records, conditions)
}

// BulkDelete deleta múltiplos registros
func (m *ModelHandler) BulkDelete(ctx context.Context, ids []interface{}) error {
	bulkOp := bulk.NewBulkOperation(m.session.dialect, m.mapping, 1000)
	return bulkOp.BulkDelete(ctx, m.session.db, ids)
}

// Preload carrega relacionamentos
func (m *ModelHandler) Preload(fields ...string) *ModelHandler {
	m.preloadFields = append(m.preloadFields, fields...)
	return m
}

// loadRelations carrega os relacionamentos solicitados
func (m *ModelHandler) loadRelations(ctx context.Context, dest interface{}) error {
	for _, field := range m.preloadFields {
		if relation, ok := m.session.relations.GetRelation(m.model, field); ok {
			if err := m.loadRelation(ctx, dest, field, relation); err != nil {
				return err
			}
		}
	}
	return nil
}

// loadRelation carrega um relacionamento específico
func (m *ModelHandler) loadRelation(ctx context.Context, dest interface{}, field string, relation Relation) error {
	// Implementa a lógica de carregamento baseada no tipo de relacionamento
	switch relation.Type {
	case relation.OneToOne:
		return m.loadOneToOne(ctx, dest, field, relation)
	case relation.OneToMany:
		return m.loadOneToMany(ctx, dest, field, relation)
	case relation.ManyToMany:
		return m.loadManyToMany(ctx, dest, field, relation)
	default:
		return fmt.Errorf("tipo de relacionamento não suportado")
	}
}

// Scope adiciona um scope à query
func (m *ModelHandler) Scope(scopes ...scope.Scope) *ModelHandler {
	m.scopes = append(m.scopes, scopes...)
	return m
}

// Paginate habilita a paginação
func (m *ModelHandler) Paginate(page, perPage int) *ModelHandler {
	m.paginator = pagination.NewPaginator(page, perPage)
	return m
}

// GetPagination retorna informações sobre a paginação
func (m *ModelHandler) GetPagination() *pagination.PageInfo {
	if m.paginator == nil {
		return nil
	}
	info := m.paginator.GetInfo()
	return &info
} 