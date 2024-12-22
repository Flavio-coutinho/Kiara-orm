package session

import (
	"context"
	"database/sql"
	
	"github.com/Flavio-coutinho/Kiara-orm/dialect"
	"github.com/Flavio-coutinho/Kiara-orm/query"
	"github.com/Flavio-coutinho/Kiara-orm/schema"
	"github.com/Flavio-coutinho/Kiara-orm/transaction"
	"github.com/Flavio-coutinho/Kiara-orm/cache"
	"github.com/Flavio-coutinho/Kiara-orm/hooks"
	"github.com/Flavio-coutinho/Kiara-orm/logger"
	"github.com/Flavio-coutinho/Kiara-orm/validator"
	"github.com/Flavio-coutinho/Kiara-orm/relation"
	"github.com/Flavio-coutinho/Kiara-orm/metrics"
)

// Session representa uma sessão de banco de dados
type Session struct {
	db        *sql.DB
	dialect   dialect.Dialect
	tx        *sql.Tx
	migrator  *schema.Migrator
	txManager *transaction.TxManager
	cache     *cache.Cache
	hooks     *hooks.HookManager
	logger    logger.Logger
	validator *validator.Validator
	relations *relation.RelationManager
	metrics *metrics.Collector
}

// NewSession cria uma nova sessão
func NewSession(db *sql.DB, dialect dialect.Dialect) *Session {
	session := &Session{
		db:        db,
		dialect:   dialect,
		migrator:  schema.NewMigrator(db, dialect),
		txManager: transaction.NewTxManager(db),
		cache:     cache.NewCache(),
		hooks:     hooks.NewHookManager(),
		logger:    logger.NewDefaultLogger(logger.INFO),
		validator: validator.NewValidator(),
		relations: relation.NewRelationManager(),
		metrics: metrics.NewCollector(),
	}
	
	// Adiciona exportador Prometheus por padrão
	session.metrics.AddExporter(metrics.NewPrometheusExporter())
	
	return session
}

// Query cria um novo query builder
func (s *Session) Query() *query.Builder {
	return query.NewBuilder(s.dialect)
}

// Exec cria um novo executor
func (s *Session) Exec(builder *query.Builder) *query.Executor {
	if s.tx != nil {
		return query.NewExecutorTx(s.tx, builder)
	}
	return query.NewExecutor(s.db, builder)
}

// AutoMigrate executa migrações automáticas
func (s *Session) AutoMigrate(ctx context.Context, models ...interface{}) error {
	return s.migrator.AutoMigrate(ctx, models...)
}

// Transaction executa uma função dentro de uma transação
func (s *Session) Transaction(ctx context.Context, fn func(tx *Session) error) error {
	return s.txManager.RunInTransaction(ctx, func(sqlTx *sql.Tx) error {
		// Cria uma nova sessão com a transação
		txSession := &Session{
			db:        s.db,
			dialect:   s.dialect,
			tx:        sqlTx,
			migrator:  s.migrator,
			txManager: s.txManager,
			cache:     s.cache,
			hooks:     s.hooks,
			logger:    s.logger,
			validator: s.validator,
			relations: s.relations,
			metrics:   s.metrics,
		}
		
		return fn(txSession)
	})
}

// Model cria um novo model handler para uma struct específica
func (s *Session) Model(model interface{}) *ModelHandler {
	return NewModelHandler(s, model)
}

// RegisterHook registra um novo hook
func (s *Session) RegisterHook(hookType hooks.HookType, hook hooks.Hook) {
	s.hooks.Register(hookType, hook)
}

// Cache retorna o gerenciador de cache
func (s *Session) Cache() *cache.Cache {
	return s.cache
}

// SetLogger define o logger
func (s *Session) SetLogger(logger logger.Logger) {
	s.logger = logger
}

// Logger retorna o logger
func (s *Session) Logger() logger.Logger {
	return s.logger
}

// Validator retorna o validador
func (s *Session) Validator() *validator.Validator {
	return s.validator
}

// HasOne define um relacionamento um-para-um
func (s *Session) HasOne(model interface{}, field string, related interface{}, foreignKey string) {
	s.relations.HasOne(model, field, related, foreignKey)
}

// HasMany define um relacionamento um-para-muitos
func (s *Session) HasMany(model interface{}, field string, related interface{}, foreignKey string) {
	s.relations.HasMany(model, field, related, foreignKey)
}

// ManyToMany define um relacionamento muitos-para-muitos
func (s *Session) ManyToMany(model interface{}, field string, related interface{}, joinTable string) {
	s.relations.ManyToMany(model, field, related, joinTable)
}

// EnablePreload habilita o carregamento automático de um relacionamento
func (s *Session) EnablePreload(model interface{}, field string) {
	s.relations.EnablePreload(model, field)
}

// Metrics retorna o coletor de métricas
func (s *Session) Metrics() *metrics.Collector {
	return s.metrics
}
 